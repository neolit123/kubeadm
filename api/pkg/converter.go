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
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

// Converter ...
type Converter struct {
	group         string
	versionKinds  []VersionKinds
	cache         map[string]Kind
	unmarshalFunc func([]byte, interface{}) error
	marshalFunc   func(interface{}) ([]byte, error)
}

// NewConverter ...
func NewConverter(group string, versionKinds []VersionKinds) *Converter {
	return &Converter{
		group:        group,
		versionKinds: versionKinds,
		cache:        map[string]Kind{},
	}
}

// GetGroup ...
func (cv *Converter) GetGroup() string {
	return cv.group
}

// AddToCache ...
func (cv *Converter) AddToCache(kind Kind) {
	key := fmt.Sprintf("%s.%s", kind.Version(), kind.Name())
	cv.cache[key] = cv.DeepCopy(nil, kind)
}

// GetFromCache ...
func (cv *Converter) GetFromCache(kind Kind) Kind {
	key := fmt.Sprintf("%s.%s", kind.Version(), kind.Name())
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

	if err := cv.unmarshalFunc(input, kind); err != nil {
		return nil, err
	}

	return kind, nil
}

// GetObject ...
func (cv *Converter) GetObject(typemeta *metav1.TypeMeta) (Kind, error) {
	gv := strings.Split(typemeta.APIVersion, "/")
	if len(gv) != 2 {
		return nil, errors.Errorf("malformed group/version: %s", typemeta.APIVersion)
	}

	for _, vk := range cv.versionKinds {
		if gv[1] != vk.Version {
			continue
		}
		for _, k := range vk.Kinds {
			if typemeta.Kind != k.Name() {
				continue
			}
			return cv.NewObject(k), nil
		}
	}
	return nil, errors.Errorf("no object for: %+v", typemeta)
}

// NewObject ...
func (cv *Converter) NewObject(kind Kind) Kind {
	t := reflect.TypeOf(kind)
	k := (reflect.New(t.Elem()).Interface()).(Kind)
	cv.SetTypeMeta(k)
	return k
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
	if dst != nil {
		cv.SetTypeMeta(dst)
	}
	return dst
}

// ConvertTo ...
func (cv *Converter) ConvertTo(in Kind, targetVersion string) (Kind, error) {
	if len(cv.versionKinds) == 0 {
		return nil, errors.New("no versions to convert to in scheme")
	}

	version := in.Version()
	kind := in.Name()

	targetVersionIdx := -1
	for i, vk := range cv.versionKinds {
		if targetVersion == vk.Version {
			targetVersionIdx = i
			break
		}
	}

	if targetVersionIdx == -1 {
		return nil, errors.Errorf("unknown target version %s", targetVersion)
	}

	versionIdx := -1
	for i := 0; i < len(cv.versionKinds); i++ {
		vk := cv.versionKinds[i]
		if version == vk.Version {
			versionIdx = i
			break
		}
	}
	if versionIdx == -1 {
		return nil, errors.Errorf("unknown version %s", version)
	}

	// already target version
	if versionIdx == targetVersionIdx {
		return in, nil
	}

	var out = in
	var err error

	if versionIdx < targetVersionIdx {
		// fmt.Println("convert up")
		goto convertUp
	}

	// fmt.Printf("versionIdx: %d, targetVersionIdx: %d\n", versionIdx, targetVersionIdx)
	// fmt.Println("convert down")

	for i := versionIdx; i > targetVersionIdx; i-- {
		vk := cv.versionKinds[i]

		for _, k := range vk.Kinds {
			if k.ConvertDownName() == kind {
				// fmt.Printf("version: %#v\n", k.Version())
				out, err = k.ConvertDown(cv, in)
				if err != nil {
					return nil, errors.Wrapf(err, "cannot convert %s/%s to %s/%s", in.Version(), in.Name(), vk.Version, k.Name())
				}
				// fmt.Printf("out: %#v\n", out)
				in = out
				kind = k.ConvertUpName()
			}
		}
	}
	return out, nil
convertUp:
	for i := versionIdx + 1; i < targetVersionIdx+1; i++ {
		vk := cv.versionKinds[i]

		for _, k := range vk.Kinds {
			if k.ConvertUpName() == kind {
				out, err = k.ConvertUp(cv, in)
				if err != nil {
					return nil, errors.Wrapf(err, "cannot convert %s/%s to %s/%s", in.Version(), in.Name(), vk.Version, k.Name())
				}
				in = out
				kind = k.ConvertDownName()
			}
		}
	}
	return out, nil
}

// ConvertToLatest ...
func (cv *Converter) ConvertToLatest(in Kind) (Kind, error) {
	if len(cv.versionKinds) == 0 {
		return nil, errors.New("no versions to convert to in scheme")
	}
	latest := cv.versionKinds[len(cv.versionKinds)-1]
	return cv.ConvertTo(in, latest.Version)
}

// GetTypeMetaFromBytes ...
func (cv *Converter) GetTypeMetaFromBytes(input []byte) (*metav1.TypeMeta, error) {
	typemeta := &metav1.TypeMeta{}
	if cv.unmarshalFunc == nil {
		return nil, errors.New("unmarshal function not set")
	}
	if err := cv.unmarshalFunc(input, typemeta); err != nil {
		return nil, errors.Wrap(err, "cannot get TypeMeta")
	}
	return typemeta, nil
}

// SetTypeMeta ...
func (cv *Converter) SetTypeMeta(kind Kind) {
	typemeta := kind.GetTypeMeta()
	typemeta.APIVersion = cv.group + "/" + kind.Version()
	typemeta.Kind = kind.Name()
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
			return nil, err
		}
		if len(doc) == 0 {
			continue
		}
		docs = append(docs, doc)
	}
	return docs, nil
}
