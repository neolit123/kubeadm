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
	"reflect"
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

// AddToCache ...
func (cv *Converter) AddToCache(kinds ...Kind) {
	for _, kind := range kinds {
		cv.cache[stringFromKind(kind)] = DeepCopy(nil, kind)
	}
}

// GetFromCache ...
func (cv *Converter) GetFromCache(kind Kind) Kind {
	cached, ok := cv.cache[stringFromKind(kind)]
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

// GetObjectFromBytes ...
func (cv *Converter) GetObjectFromBytes(typemeta *metav1.TypeMeta, input []byte) (Kind, error) {
	kind, err := cv.GetObject(typemeta)
	if err != nil {
		return nil, err
	}
	if cv.unmarshalFunc == nil {
		return nil, errors.New("unmarshal function not set")
	}
	if err := cv.unmarshalFunc(input, kind); err != nil {
		return nil, err
	}
	return kind, nil
}

// GetObject ...
func (cv *Converter) GetObject(typemeta *metav1.TypeMeta) (Kind, error) {
	gvk := typemeta.GroupVersionKind()
	for _, g := range cv.groups {
		if g.Name != gvk.Group {
			continue
		}
		for _, vk := range g.Versions {
			if gvk.Version != vk.Version {
				continue
			}
			for _, k := range vk.Kinds {
				if gvk.Kind != k.GetDefaultTypeMeta().Kind {
					continue
				}
				new := NewObject(k)
				SetDefaultTypeMeta(new)
				return new, nil
			}
		}
	}
	return nil, errors.Errorf("no object for %s", typemeta)
}

// getGroup ...
func (cv *Converter) getGroup(name string) (*Group, int, error) {
	if len(cv.groups) == 0 {
		return nil, -1, errors.New("no groups defined")
	}
	for i := range cv.groups {
		g := cv.groups[i]
		if name == g.Name {
			return &g, i, nil
		}
	}
	return nil, -1, errors.Errorf("unknown group %q", name)
}

// ConvertTo ...
func (cv *Converter) ConvertTo(in *KindSpec, targetGroup, targetVersion string) (*KindSpec, error) {
	if err := ValidateKindSpec(in); err != nil {
		return nil, err
	}
	if len(in.Kinds) == 0 {
		return nil, errors.New("empty input spec")
	}

	_, targetGroupIdx, err := cv.getGroup(targetGroup)
	if err != nil {
		return nil, err
	}
	sourceGVK := in.Kinds[0].GetDefaultTypeMeta().GroupVersionKind()
	sourceGroupObj, sourceGroupIdx, err := cv.getGroup(sourceGVK.Group)
	if err != nil {
		return nil, err
	}

	var start, end, step int
	var convertUp bool
	var convertString string

	// flatten
	kinds := []Kind{}
	for _, g := range cv.groups {
		for _, vk := range g.Versions {
			for _, k := range vk.Kinds {
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
			convertFunc = k.ConvertUp
			convertSpecFunc = k.ConvertUpSpec
		} else {
			convertFunc = k.ConvertDown
			convertSpecFunc = k.ConvertDownSpec
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
		"last matching spec: %s, last requested spec: %s, last called convert on: %v",
		targetGroup, targetVersion, convertString, originalInput, lastSpec, in, reflect.TypeOf(lastKind))
}

// ConvertToLatest ...
func (cv *Converter) ConvertToLatest(in *KindSpec, targetGroup string) (*KindSpec, error) {
	g, _, err := cv.getGroup(targetGroup)
	if err != nil {
		return nil, err
	}
	latest := g.Versions[len(g.Versions)-1]
	return cv.ConvertTo(in, targetGroup, latest.Version)
}

// ConvertToOldest ...
func (cv *Converter) ConvertToOldest(in *KindSpec, targetGroup string) (*KindSpec, error) {
	g, _, err := cv.getGroup(targetGroup)
	if err != nil {
		return nil, err
	}
	oldest := g.Versions[0]
	return cv.ConvertTo(in, targetGroup, oldest.Version)
}

// TypeMetaFromBytes ...
func (cv *Converter) TypeMetaFromBytes(input []byte) (*metav1.TypeMeta, error) {
	typemeta := &metav1.TypeMeta{}
	if cv.unmarshalFunc == nil {
		return nil, errors.New("unmarshal function not set")
	}
	if err := cv.unmarshalFunc(input, typemeta); err != nil {
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

// GetAnnotations ...
func (cv *Converter) GetAnnotations(in []byte) (map[string]string, error) {
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
		typemeta, err := cv.TypeMetaFromBytes([]byte(v))
		if err != nil {
			return err
		}
		keyFromAnnotation := strings.TrimPrefix(k, ConverterCacheAnnotation+".")
		// in case typemeta is not present in the serialized object attempt to use the annotation key
		if len(typemeta.APIVersion) == 0 && len(typemeta.Kind) == 0 {
			typemeta = typemetaFromString(keyFromAnnotation)
			if typemeta == nil {
				return errors.Errorf("cannot parse typemeta from annotation key %q", keyFromAnnotation)
			}
		}
		kind, err := cv.GetObjectFromBytes(typemeta, []byte(v))
		if err != nil {
			return err
		}
		keyFromKind := stringFromKind(kind)
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
