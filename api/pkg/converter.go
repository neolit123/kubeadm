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
	"reflect"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewConverter ...
func NewConverter(groups []Group) (*Converter, error) {
	if err := ValidateGroups(groups); err != nil {
		return nil, err
	}
	return &Converter{
		groups: groups,
		cache:  map[string]Kind{},
	}, nil
}

// ValidateGroups ...
func ValidateGroups(groups []Group) error {
	if len(groups) == 0 {
		return errors.New("found an empty or nil list of groups")
	}
	for _, g := range groups {
		if len(g.Name) == 0 {
			return errors.New("found an empty group name")
		}
		for _, vk := range g.Versions {
			if len(vk.Version) == 0 {
				return errors.Errorf("group %q has a version with empty name", g.Name)
			}
			for _, k := range vk.Kinds {
				t := reflect.TypeOf(k)
				gvk := k.GetDefaultTypeMeta().GroupVersionKind()
				if gvk.Group != g.Name {
					return errors.Errorf("expected group for object %v: %q, got: %q", t, g.Name, gvk.Group)
				}
				if gvk.Version != vk.Version {
					return errors.Errorf("expected version for object %v: %q, got: %q", t, vk.Version, gvk.Version)
				}
				if gvk.Kind == "" {
					return errors.Errorf("empty Kind for object %v", t)
				}
				if err := ValidateKindSpec(k.ConvertUpSpec()); err != nil {
					return errors.Wrapf(err, "error in ConvertUpSpec for %v", t)
				}
				if err := ValidateKindSpec(k.ConvertDownSpec()); err != nil {
					return errors.Wrapf(err, "error in ConvertDownSpec for %v", t)
				}
			}
		}
	}
	return nil
}

// ValidateKindSpec ...
func ValidateKindSpec(in *KindSpec) error {
	if in == nil {
		return errors.New("nil spec")
	}
	var groupVersion string
	for i, k := range in.Kinds {
		tm := k.GetDefaultTypeMeta()
		if tm.APIVersion == "" {
			return errors.Errorf("object with empty APIVersion at position %d", i)
		}
		if tm.Kind == "" {
			return errors.Errorf("object with empty Kind at position %d", i)
		}
		if groupVersion == "" {
			groupVersion = tm.APIVersion
			continue
		}
		if groupVersion != tm.APIVersion {
			return errors.Errorf("found multiple APIVersions")
		}
	}
	return nil
}

// AddToCache ...
func (cv *Converter) AddToCache(kind Kind) {
	key := kind.GetDefaultTypeMeta().String()
	cv.cache[key] = cv.DeepCopy(nil, kind)
}

// GetFromCache ...
func (cv *Converter) GetFromCache(kind Kind) Kind {
	key := kind.GetDefaultTypeMeta().String()
	cached, ok := cv.cache[key]
	if !ok {
		return nil
	}
	return cv.DeepCopy(nil, cached)
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
				new := cv.NewObject(k)
				cv.SetDefaultTypeMeta(new)
				return new, nil
			}
		}
	}
	return nil, errors.Errorf("no object for: %+v", typemeta)
}

// NewObject ...
func (cv *Converter) NewObject(kind Kind) Kind {
	t := reflect.TypeOf(kind)
	return (reflect.New(t.Elem()).Interface()).(Kind)
}

// DeepCopy ...
func (cv *Converter) DeepCopy(dst Kind, src Kind) Kind {
	if src == nil {
		panic("nil value passed to DeepCopy")
	}
	bytes, err := cv.Marshal(src)
	if err != nil {
		panic("error marshal: " + err.Error())
	}
	if dst == nil {
		dst = cv.NewObject(src)
	}
	if err := cv.Unmarshal(bytes, dst); err != nil {
		panic("error unmarshal: " + err.Error())
	}
	cv.SetDefaultTypeMeta(dst)
	return dst
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
	var convertSpecString string

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
	convertSpecString = "ConvertUpSpec"
	goto convert

covertDown:
	start = len(kinds) - 1
	end = -1
	step = -1
	convertSpecString = "ConvertDownSpec"
	convertUp = false

convert:
	var out *KindSpec
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
		if !in.EqualKinds(convertSpecFunc()) {
			continue
		}
		out, err = convertFunc(cv, in)
		if err != nil {
			return nil, err
		}
		gvk := out.Kinds[0].GetDefaultTypeMeta().GroupVersionKind()
		if gvk.Group == targetGroup && gvk.Version == targetVersion {
			return out, nil
		}
		in = out
	}
	return nil, errors.Errorf("no matching %s for %s", convertSpecString, in)
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

// SetDefaultTypeMeta ...
func (cv *Converter) SetDefaultTypeMeta(kind Kind) {
	typemeta := kind.GetTypeMeta()
	*typemeta = *kind.GetDefaultTypeMeta()
}

// SetMarshalFunc ...
func (cv *Converter) SetMarshalFunc(f func(interface{}) ([]byte, error)) {
	cv.marshalFunc = f
}

// SetUnmarshalFunc ...
func (cv *Converter) SetUnmarshalFunc(f func([]byte, interface{}) error) {
	cv.unmarshalFunc = f
}

// Marshal ...
func (cv *Converter) Marshal(k Kind) ([]byte, error) {
	if cv.marshalFunc == nil {
		return nil, errors.New("marshal function not set")
	}
	return cv.marshalFunc(k)
}

// Unmarshal ...
func (cv *Converter) Unmarshal(b []byte, k Kind) error {
	if cv.unmarshalFunc == nil {
		return errors.New("unmarshal function not set")
	}
	return cv.unmarshalFunc(b, k)
}

// String ...
func (s *KindSpec) String() string {
	str := "KindSpec{ "
	str += s.kindsString()
	return str
}

// kindsString ...
func (s *KindSpec) kindsString() string {
	str := ""
	for _, k := range s.Kinds {
		if k == nil {
			str += "nil "
			continue
		}
		str += k.GetDefaultTypeMeta().String() + " "
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

// NewKindSpec ...
func NewKindSpec() *KindSpec {
	return &KindSpec{}
}
