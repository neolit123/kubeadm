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
				gvk := k.GetDefaultTypeMeta().GroupVersionKind()
				if gvk.Group != g.Name {
					return errors.Errorf("expected group for object %v: %q, got: %q", reflect.TypeOf(k), g.Name, gvk.Group)
				}
				if gvk.Version != vk.Version {
					return errors.Errorf("expected version for object %v: %q, got: %q", reflect.TypeOf(k), vk.Version, gvk.Version)
				}
			}
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
	if in == nil {
		return nil, errors.New("ConvertTo received a nil KindSpec")
	}
	if len(in.Kinds) == 0 {
		return nil, errors.New("ConvertTo received an empty list of Kinds")
	}

	// get the source and target groups
	targetGroupObj, targetGroupIdx, err := cv.getGroup(targetGroup)
	if err != nil {
		return nil, err
	}
	sourceGroup := in.Kinds[0].GetDefaultTypeMeta().GroupVersionKind().Group
	sourceGroupObj, sourceGroupIdx, err := cv.getGroup(sourceGroup)
	if err != nil {
		return nil, err
	}

	var versionKinds = targetGroupObj.Versions
	if targetGroupIdx == sourceGroupIdx {
		goto convertSameGroup
	} else if targetGroupIdx < sourceGroupIdx {
		versionKinds = sourceGroupObj.Versions
	}

	for _, vk := range versionKinds {
		for _, k := range vk.Kinds {
			if sourceGroupIdx < targetGroupIdx {
				if in.EqualKinds(k.ConvertUpSpec()) {
					out, err := k.ConvertUp(cv, in)
					if err != nil {
						return nil, errors.Wrapf(err, "ConvertUp for %s cannot convert %s", k.GetDefaultTypeMeta(), in)
					}
					if out == nil {
						return nil, errors.Wrapf(err, "ConvertUp for %s returned nil", k.GetDefaultTypeMeta())
					}
					if len(out.Kinds) == 0 {
						return nil, errors.Wrapf(err, "ConvertUp for %s returned an empty list of Kinds", k.GetDefaultTypeMeta())
					}
					return out, nil
				}
			} else {
				if in.EqualKinds(k.ConvertDownSpec()) {
					out, err := k.ConvertDown(cv, in)
					if err != nil {
						return nil, errors.Wrapf(err, "ConvertDown for %s cannot convert %s", k.GetDefaultTypeMeta(), in)
					}
					if out == nil {
						return nil, errors.Wrapf(err, "ConvertDown for %s returned nil", k.GetDefaultTypeMeta())
					}
					if len(out.Kinds) == 0 {
						return nil, errors.Wrapf(err, "ConvertDown for %s returned an empty list of Kinds", k.GetDefaultTypeMeta())
					}
					return out, nil
				}
			}
		}
	}

convertSameGroup:
	// get the current version index
	versionIdx := -1
	for i := 0; i < len(versionKinds); i++ {
		vk := versionKinds[i]
		for _, k := range vk.Kinds {
			if in.EqualKinds(k.ConvertDownSpec()) {
				versionIdx = i
				goto breakLoop
			}
		}
	}
breakLoop:
	if versionIdx == -1 {
		return nil, errors.Errorf("cannot find %s in group %q", in, targetGroup)
	}

	// get the target version index
	targetVersionIdx := -1
	for i, vk := range versionKinds {
		if targetVersion == vk.Version {
			targetVersionIdx = i
			break
		}
	}
	if targetVersionIdx == -1 {
		return nil, errors.Errorf("unknown target version %s", targetVersion)
	}

	// already target version
	if versionIdx == targetVersionIdx {
		return in, nil
	}

	var out = in
	if versionIdx < targetVersionIdx {
		goto convertUp
	}

	// To convert down, iterate from the current version until the target version
	// is reached (not including).
	for i := versionIdx; i > targetVersionIdx; i-- {
		vk := versionKinds[i]
		found := false
		for _, k := range vk.Kinds {
			if in.EqualKinds(k.ConvertDownSpec()) {
				out, err = k.ConvertDown(cv, in)
				if err != nil {
					return nil, errors.Wrapf(err, "ConvertDown for %s cannot convert %s", k.GetDefaultTypeMeta(), in)
				}
				if out == nil {
					return nil, errors.Wrapf(err, "ConvertDown for %s returned nil", k.GetDefaultTypeMeta())
				}
				if len(out.Kinds) == 0 {
					return nil, errors.Wrapf(err, "ConvertDown for %s returned an empty list of Kinds", k.GetDefaultTypeMeta())
				}
				in = out
				found = true
				break
			}
		}
		if !found {
			return nil, errors.Errorf("cannot down convert %s, no matching ConvertDownSpec() in version %q", in, vk.Version)
		}
	}
	return out, nil

	// To convert up, iterate from the current version index + 1 (next version)
	// until the target version is reached (including).
convertUp:
	for i := versionIdx + 1; i < targetVersionIdx+1; i++ {
		vk := versionKinds[i]
		found := false
		for _, k := range vk.Kinds {
			if in.EqualKinds(k.ConvertUpSpec()) {
				out, err = k.ConvertUp(cv, in)
				if err != nil {
					return nil, errors.Wrapf(err, "ConvertUp for %s cannot convert %s", k.GetDefaultTypeMeta(), in)
				}
				if out == nil {
					return nil, errors.Wrapf(err, "ConvertUp for %s returned nil", k.GetDefaultTypeMeta())
				}
				if len(out.Kinds) == 0 {
					return nil, errors.Wrapf(err, "ConvertUp for %s returned an empty list of Kinds", k.GetDefaultTypeMeta())
				}
				in = out
				found = true
				break
			}
		}
		if !found {
			return nil, errors.Errorf("cannot up convert %s, no matching ConvertUpSpec() in version %q", in, vk.Version)
		}
	}
	return out, nil
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
