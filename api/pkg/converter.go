/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewConverter ...
func NewConverter() *Converter {
	return &Converter{
		output:        ioutil.Discard,
		cache:         map[string]Kind{},
		marshalFunc:   json.Marshal,
		unmarshalFunc: json.Unmarshal,
	}
}

// WithGroups ...
func (cv *Converter) WithGroups(groups []Group) *Converter {
	cv.groups = groups
	return cv
}

// WithOutput ...
func (cv *Converter) WithOutput(output io.Writer) *Converter {
	cv.output = output
	return cv
}

// WithCache ...
func (cv *Converter) WithCache(cache map[string]Kind) *Converter {
	cv.cache = cache
	return cv
}

// WithMarshalFunc ...
func (cv *Converter) WithMarshalFunc(f func(interface{}) ([]byte, error)) *Converter {
	cv.marshalFunc = f
	return cv
}

// WithUnmarshalFunc ...
func (cv *Converter) WithUnmarshalFunc(f func([]byte, interface{}) error) *Converter {
	cv.unmarshalFunc = f
	return cv
}

// AddKindsToCache ...
func (cv *Converter) AddKindsToCache(kinds ...Kind) {
	for _, kind := range kinds {
		cv.cache[gvkStringFromKind(kind)] = DeepCopy(nil, kind)
	}
}

// KindFromCache ...
func (cv *Converter) KindFromCache(kind Kind) Kind {
	cached, ok := cv.cache[gvkStringFromKind(kind)]
	if !ok {
		return nil
	}
	return DeepCopy(nil, cached)
}

// ClearCache ...
func (cv *Converter) ClearCache() {
	for k := range cv.cache {
		delete(cv.cache, k)
	}
}

// ReadKind ...
func (cv *Converter) ReadKind(typemeta *metav1.TypeMeta, input []byte) (Kind, error) {
	var err error
	if typemeta == nil {
		typemeta, err = cv.ReadTypeMeta(input)
		if err != nil {
			return nil, err
		}
	}
	kind, err := cv.NewKindInstance(typemeta)
	if err != nil {
		return nil, err
	}
	if err := cv.Unmarshal(input, kind); err != nil {
		return nil, err
	}
	return kind, nil
}

// NewKindInstance ...
func (cv *Converter) NewKindInstance(typemeta *metav1.TypeMeta) (Kind, error) {
	gvk := typemeta.GroupVersionKind()
	for _, g := range cv.groups {
		if g.Group != gvk.Group {
			continue
		}
		for _, v := range g.Versions {
			if gvk.Version != v.Version {
				continue
			}
			for _, k := range v.Kinds {
				if gvk.Kind != k.GetDefaultTypeMeta().Kind {
					continue
				}
				new := newKindInstance(k)
				SetDefaultTypeMeta(new)
				return new, nil
			}
		}
	}
	return nil, errors.Errorf("no object for %s", typemeta)
}

func (cv *Converter) convertFromHub(in *KindSpec, targetGroupObj *Group, targetVersion, errorPrefix string) (*KindSpec, error) {
	for _, v := range targetGroupObj.Versions {
		for _, k := range v.Kinds {
			from := k.ConvertFromSpec()
			if !from.EqualKinds(in) {
				continue
			}
			to := k.ConvertToSpec()
			gvk := to.Kinds[0].GetDefaultTypeMeta().GroupVersionKind()
			if gvk.Version == targetVersion {
				out, err := k.ConvertTo(cv, in)
				if err != nil {
					return nil, errors.Wrapf(err, "%s: ConvertFrom() of %T failed", errorPrefix, k)
				}
				if len(out.Kinds) == 0 {
					return nil, errors.Errorf("%s: ConvertFrom() of %T returned empty spec", errorPrefix, k)
				}
				return out, nil
			}
		}
	}
	return nil, errors.Errorf("%s: could not convert from the hub spec %s", errorPrefix, in)
}

func (cv *Converter) convertToHub(in *KindSpec, targetGroupObj *Group, errorPrefix string) (*KindSpec, error) {
	for _, v := range targetGroupObj.Versions {
		for _, k := range v.Kinds {
			to := k.ConvertToSpec()
			if !to.EqualKinds(in) {
				continue
			}
			out, err := k.ConvertTo(cv, in)
			if err != nil {
				return nil, errors.Wrapf(err, "%s: ConvertTo() of %T failed", errorPrefix, k)
			}
			if len(out.Kinds) == 0 {
				return nil, errors.Errorf("%s: ConvertTo() of %T returned empty spec", errorPrefix, k)
			}
			return out, nil
		}
	}
	return nil, errors.Errorf("%s: could not convert from the hub spec %s", errorPrefix, in)
}

// ConvertTo ...
func (cv *Converter) ConvertTo(in *KindSpec, targetGroup, targetVersion string) (*KindSpec, error) {
	if err := ValidateKindSpec(in); err != nil {
		return nil, err
	}
	if len(in.Kinds) == 0 {
		return nil, errors.New("empty input spec")
	}

	targetGroupObj, targetGroupIdx, err := getGroup(cv.groups, targetGroup)
	if err != nil {
		return nil, err
	}
	sourceGVK := in.Kinds[0].GetDefaultTypeMeta().GroupVersionKind()
	sourceGroupObj, sourceGroupIdx, err := getGroup(cv.groups, sourceGVK.Group)
	if err != nil {
		return nil, err
	}

	var errorPrefix = fmt.Sprintf("the converter did not reach %s/%s for input %s",
		targetGroup, targetVersion, in)

	var hubSpec, out *KindSpec
	if sourceGroupIdx != targetGroupIdx {
		goto linearVersions
	}

	for _, v := range targetGroupObj.Versions {
		for _, k := range v.Kinds {
			from := k.ConvertFromSpec()
			if len(from.Kinds) == 0 {
				continue
			}
			if hubSpec != nil && !from.EqualKinds(hubSpec) {
				goto linearVersions
			}
			hubSpec = from
		}
	}

	if in.EqualKinds(hubSpec) {
		out, err = cv.convertFromHub(in, targetGroupObj, targetVersion, errorPrefix)
		if err != nil {
			return nil, err
		}
		return out, nil
	} else {
		out, err = cv.convertToHub(in, targetGroupObj, errorPrefix)
		if err != nil {
			return nil, err
		}
		out, err = cv.convertFromHub(out, targetGroupObj, targetVersion, errorPrefix)
		if err != nil {
			return nil, err
		}
		return out, nil
	}

linearVersions:
	var start, end, step int
	var convertUp bool
	var convertString string

	// flatten
	kinds := []Kind{}
	for _, g := range cv.groups {
		for _, v := range g.Versions {
			for _, k := range v.Kinds {
				kinds = append(kinds, k)
			}
		}
	}

	if sourceGroupIdx < targetGroupIdx {
		goto covertUp
	} else if sourceGroupIdx > targetGroupIdx {
		goto covertDown
	} else {
		sourceVersionIdx := -1
		targetVersionIdx := -1
		for i, v := range sourceGroupObj.Versions {
			if sourceGVK.Version == v.Version {
				sourceVersionIdx = i
			}
			if targetVersion == v.Version {
				targetVersionIdx = i
			}
		}
		if sourceVersionIdx < targetVersionIdx {
			goto covertUp
		} else if sourceVersionIdx > targetVersionIdx {
			goto covertDown
		} else {
			// nothing to do
			return in, nil
		}
	}

covertUp:
	start = 0
	end = len(kinds)
	step = 1
	convertUp = true
	convertString = "converting up"
	goto convert

covertDown:
	start = len(kinds) - 1
	end = -1
	step = -1
	convertString = "converting down"
	convertUp = false

convert:
	var originalInput = in
	var lastKind Kind
	var lastSpec *KindSpec
	var convertFunc func(*Converter, *KindSpec) (*KindSpec, error)
	var convertSpecFunc func() *KindSpec
	for i := start; i != end; i += step {
		k := kinds[i]
		if convertUp {
			convertFunc = k.ConvertFrom
			convertSpecFunc = k.ConvertFromSpec
		} else {
			convertFunc = k.ConvertTo
			convertSpecFunc = k.ConvertToSpec
		}
		spec := convertSpecFunc()
		if !in.EqualKinds(spec) {
			continue
		}
		lastSpec = spec
		in, err = convertFunc(cv, in)
		if err != nil {
			return nil, err
		}
		lastKind = k
		gvk := in.Kinds[0].GetDefaultTypeMeta().GroupVersionKind()
		if gvk.Group == targetGroup && gvk.Version == targetVersion {
			return in, nil
		}
	}
	return nil, errors.Errorf("the converter did not reach %s/%s when %s for input %s: "+
		"last matching spec: %s, last requested spec: %s, last called convert on %T",
		targetGroup, targetVersion, convertString, originalInput, lastSpec, in, lastKind)
}

// ConvertToLatest ...
func (cv *Converter) ConvertToLatest(in *KindSpec, targetGroup string) (*KindSpec, error) {
	g, _, err := getGroup(cv.groups, targetGroup)
	if err != nil {
		return nil, err
	}
	latest := g.Versions[len(g.Versions)-1]
	return cv.ConvertTo(in, targetGroup, latest.Version)
}

// ConvertToOldest ...
func (cv *Converter) ConvertToOldest(in *KindSpec, targetGroup string) (*KindSpec, error) {
	g, _, err := getGroup(cv.groups, targetGroup)
	if err != nil {
		return nil, err
	}
	oldest := g.Versions[0]
	return cv.ConvertTo(in, targetGroup, oldest.Version)
}

// ConvertToPreferred ...
func (cv *Converter) ConvertToPreferred(in *KindSpec, targetGroup string) (*KindSpec, error) {
	pref, err := GetPreferredVersion(cv.groups, targetGroup)
	if err != nil {
		return nil, err
	}
	return cv.ConvertTo(in, targetGroup, pref.Version)
}

// ReadTypeMeta ...
func (cv *Converter) ReadTypeMeta(input []byte) (*metav1.TypeMeta, error) {
	typemeta := &metav1.TypeMeta{}
	if err := cv.Unmarshal(input, typemeta); err != nil {
		return nil, errors.Wrap(err, "cannot get TypeMeta")
	}
	return typemeta, nil
}

// Marshal ...
func (cv *Converter) Marshal(k interface{}) ([]byte, error) {
	if cv.marshalFunc == nil {
		return nil, errors.New("marshal function not set")
	}
	return cv.marshalFunc(k)
}

// Unmarshal ...
func (cv *Converter) Unmarshal(b []byte, k interface{}) error {
	if cv.unmarshalFunc == nil {
		return errors.New("unmarshal function not set")
	}
	return cv.unmarshalFunc(b, k)
}

// DeleteMetadata ...
func (cv *Converter) DeleteMetadata(in []byte) ([]byte, error) {
	const errStr = "could not delete metadata"
	u := map[string]interface{}{}
	if err := cv.Unmarshal(in, &u); err != nil {
		return nil, errors.Wrap(err, errStr)
	}
	delete(u, "metadata")
	bytes, err := cv.Marshal(u)
	if err != nil {
		return nil, errors.Wrap(err, errStr)
	}
	return bytes, nil
}

// ReadAnnotations ...
func (cv *Converter) ReadAnnotations(in []byte) (map[string]string, error) {
	u := map[string]interface{}{}
	if err := cv.Unmarshal(in, &u); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling")
	}
	_, annotations, err := getMetadataAnnotations(u)
	if err != nil {
		return nil, err
	}
	return annotations, nil
}

// SetAnnotations ...
func (cv *Converter) SetAnnotations(in []byte, annotations map[string]string) ([]byte, error) {
	u := map[string]interface{}{}
	if err := cv.Unmarshal(in, &u); err != nil {
		return nil, errors.Wrap(err, "error unmarshaling")
	}
	metadata, _, _ := getMetadataAnnotations(u)
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	metadata["annotations"] = annotations
	u["metadata"] = metadata
	out, err := cv.Marshal(&u)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling")
	}
	return out, nil
}

// AddCacheToAnnotations ...
func (cv *Converter) AddCacheToAnnotations(annotations map[string]string) error {
	for k, v := range cv.cache {
		key := fmt.Sprintf("%s.%s", ConverterCacheAnnotation, k)
		bytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		annotations[key] = string(bytes)
	}
	return nil
}

// AddAnnotationsToCache ...
func (cv *Converter) AddAnnotationsToCache(annotations map[string]string) error {
	for k, v := range annotations {
		if !strings.HasPrefix(k, ConverterCacheAnnotation+".") {
			continue
		}
		typemeta, err := cv.ReadTypeMeta([]byte(v))
		if err != nil {
			return err
		}
		keyFromAnnotation := strings.TrimPrefix(k, ConverterCacheAnnotation+".")
		// in case typemeta is not present in the serialized object attempt to use the annotation key
		if len(typemeta.APIVersion) == 0 && len(typemeta.Kind) == 0 {
			typemeta = typemetaFromGVKString(keyFromAnnotation)
			if typemeta == nil {
				return errors.Errorf("cannot parse typemeta from annotation key %q", keyFromAnnotation)
			}
		}
		kind, err := cv.ReadKind(typemeta, []byte(v))
		if err != nil {
			return err
		}
		keyFromKind := gvkStringFromKind(kind)
		if keyFromKind != keyFromAnnotation {
			return errors.Errorf("mismatch in annotation key for object %+v: found %q, expected %q",
				typemeta, keyFromAnnotation, keyFromKind)
		}
		cv.cache[keyFromKind] = kind
	}
	return nil
}

// NewKindSpec ...
func NewKindSpec() *KindSpec {
	return &KindSpec{}
}

// String ...
func (s *KindSpec) String() string {
	str := "KindSpec{"
	str += s.kindsString()
	str += "}"
	return str
}

// kindsString ...
func (s *KindSpec) kindsString() string {
	str := ""
	for _, k := range s.Kinds {
		if k == nil {
			str += "nil,"
			continue
		}
		str += k.GetDefaultTypeMeta().String() + ","
	}
	return str
}

// EqualKinds ...
func (s *KindSpec) EqualKinds(e *KindSpec) bool {
	return s.kindsString() == e.kindsString()
}

// WithKinds ...
func (s *KindSpec) WithKinds(in ...Kind) *KindSpec {
	s.Kinds = in
	return s
}
