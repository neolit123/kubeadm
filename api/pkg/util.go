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
	"encoding/json"
	"io"
	"reflect"

	"github.com/pkg/errors"

	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

// SplitDocuments ...
func SplitDocuments(b []byte) ([][]byte, error) {
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

// JoinDocuments ...
func JoinDocuments(docs [][]byte) []byte {
	return bytes.Join(docs, []byte("\n---\n"))
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
		for i, vk := range g.Versions {
			if len(vk.Version) == 0 {
				return errors.Errorf("group %q has a version with empty name at position %d", g.Name, i)
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

// SetDefaultTypeMeta ...
func SetDefaultTypeMeta(kind Kind) {
	typemeta := kind.GetTypeMeta()
	*typemeta = *kind.GetDefaultTypeMeta()
}

// NewObject ...
func NewObject(kind Kind) Kind {
	t := reflect.TypeOf(kind)
	return (reflect.New(t.Elem()).Interface()).(Kind)
}

// DeepCopy ...
func DeepCopy(dst Kind, src Kind) Kind {
	if src == nil {
		panic("nil value passed to DeepCopy")
	}
	bytes, err := json.Marshal(src)
	if err != nil {
		panic("error marshal: " + err.Error())
	}
	if dst == nil {
		dst = NewObject(src)
	}
	if err := json.Unmarshal(bytes, dst); err != nil {
		panic("error unmarshal: " + err.Error())
	}
	SetDefaultTypeMeta(dst)
	return dst
}
