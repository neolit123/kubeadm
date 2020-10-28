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
	"bufio"
	"bytes"
	"io"
	"reflect"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
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
	key := kind.GetTypeMeta().String()
	cv.cache[key] = cv.DeepCopy(nil, kind)
}

// GetFromCache ...
func (cv *Converter) GetFromCache(kind Kind) Kind {
	key := kind.GetTypeMeta().String()
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
				cv.SetGetDefaultTypeMeta(new)
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
	cv.SetGetDefaultTypeMeta(dst)
	return dst
}

// ConvertTo ...
// TODO: support accross groups is possible, but not implemented yet.
func (cv *Converter) ConvertTo(in *ConvertSpec, group, targetVersion string) (*ConvertSpec, error) {
	g, err := cv.getGroup(group)
	if err != nil {
		return nil, err
	}
	versionKinds := g.Versions

	// get the current version index
	versionIdx := -1
	for i := 0; i < len(versionKinds); i++ {
		vk := versionKinds[i]
		for _, k := range vk.Kinds {
			if in.Equals(k.ConvertDownSpec()) {
				versionIdx = i
				goto break_loop
			}
		}
	}
break_loop:
	if versionIdx == -1 {
		return nil, errors.Errorf("cannot find %s in group %q", in, group)
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
			if in.Equals(k.ConvertDownSpec()) {
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
			if in.Equals(k.ConvertUpSpec()) {
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

// getGroup ...
func (cv *Converter) getGroup(name string) (*Group, error) {
	if len(cv.groups) == 0 {
		return nil, errors.New("no groups defined")
	}
	for i := range cv.groups {
		g := cv.groups[i]
		if name == g.Name {
			return &g, nil
		}
	}
	return nil, errors.Errorf("unknown group %q", name)
}

// ConvertToLatest ...
func (cv *Converter) ConvertToLatest(in *ConvertSpec, group string) (*ConvertSpec, error) {
	g, err := cv.getGroup(group)
	if err != nil {
		return nil, err
	}
	latest := g.Versions[len(g.Versions)-1]
	return cv.ConvertTo(in, group, latest.Version)
}

// ConvertToOldest ...
func (cv *Converter) ConvertToOldest(in *ConvertSpec, group string) (*ConvertSpec, error) {
	g, err := cv.getGroup(group)
	if err != nil {
		return nil, err
	}
	oldest := g.Versions[0]
	return cv.ConvertTo(in, group, oldest.Version)
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

// SetGetDefaultTypeMeta ...
func (cv *Converter) SetGetDefaultTypeMeta(kind Kind) {
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

// SplitDocuments ...
func (cv *Converter) SplitDocuments(b []byte) ([][]byte, error) {
	docs := [][]byte{}
	buf := bytes.NewBuffer(b)
	reader := utilyaml.NewYAMLReader(bufio.NewReader(buf))
	for {
		doc, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, errors.Wrap(err, "could not split documents")
		}
		if len(doc) == 0 {
			continue
		}
		docs = append(docs, doc)
	}
	return docs, nil
}

// String ...
func (s *ConvertSpec) String() string {
	str := "ConverterSpec{ "
	for _, k := range s.Kinds {
		str += k.GetDefaultTypeMeta().String() + " "
	}
	str += "}"
	return str
}

// Equals ...
func (s *ConvertSpec) Equals(e *ConvertSpec) bool {
	return s.String() == e.String()
}
