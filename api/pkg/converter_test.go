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
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testGroup = "testgroup"

func TestValidateGroups(t *testing.T) {
	testCases := []struct {
		name          string
		groups        []Group
		expectedError bool
	}{
		{
			name:          "valid: passes validation",
			expectedError: false,
			groups:        []Group{Group{Name: testGroup, Versions: []VersionKinds{{Version: "v1beta1", Kinds: []Kind{&testFoo{}}}}}},
		},
		{
			name:          "invalid: unknown group",
			expectedError: true,
			groups:        []Group{Group{Name: "foo", Versions: []VersionKinds{{Version: "v1beta1", Kinds: []Kind{&testFoo{}}}}}},
		},
		{
			name:          "invalid: unknown version",
			expectedError: true,
			groups:        []Group{Group{Name: testGroup, Versions: []VersionKinds{{Version: "foo", Kinds: []Kind{&testFoo{}}}}}},
		},
		{
			name:          "invalid: empty groups",
			expectedError: true,
			groups:        []Group{},
		},
		{
			name:          "invalid: empty group name",
			expectedError: true,
			groups:        []Group{Group{Name: "", Versions: []VersionKinds{{Version: "foo", Kinds: []Kind{&testFoo{}}}}}},
		},
		{
			name:          "invalid: empty version name",
			expectedError: true,
			groups:        []Group{Group{Name: testGroup, Versions: []VersionKinds{{Version: "", Kinds: []Kind{&testFoo{}}}}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewConverter(tc.groups)
			if (err != nil) != tc.expectedError {
				t.Fatalf("expected error %v, got %v, error: %v", tc.expectedError, err != nil, err)
			}
		})
	}
}

func TestSplitDocuments(t *testing.T) {
	foo := `{ "foo": "Foo" }`
	bar := `{ "bar": "Bar" }`
	multiDoc := foo + "\n---\n" + bar

	cv := (*Converter)(nil)
	docs, err := cv.SplitDocuments([]byte(multiDoc))
	if err != nil {
		t.Fatalf("document split error: %v", err)
	}
	if len(docs) != 2 {
		t.Fatalf("expected %d documents, got %d", 2, len(docs))
	}

	expectedFoo := []byte(foo + "\n")
	if !bytes.Equal(docs[0], expectedFoo) {
		t.Fatalf("expected first document:\n%v\ngot:\n%v", expectedFoo, docs[0])
	}
	expectedBar := []byte(bar + "\n")
	if !bytes.Equal(docs[1], expectedBar) {
		t.Fatalf("expected second document:\n%s\ngot:\n%s", expectedBar, docs[1])
	}
}

func TestConvert(t *testing.T) {
	g := []Group{
		{
			Name: testGroup,
			Versions: []VersionKinds{
				{"v1beta1", []Kind{&testFoo{}}},
				{"v1beta2", []Kind{&testBar{}}},
				{"v1beta3", []Kind{&testZed{}}},
			},
		},
	}
	cv, err := NewConverter(g)
	if err != nil {
		t.Fatal(err)
	}
	cv.SetMarshalFunc(json.Marshal)
	cv.SetUnmarshalFunc(json.Unmarshal)

	testZedJSON := []byte(`{"kind": "testZed", "apiVersion": "testgroup/v1beta3", "a": "A", "b": "B", "c": "C"}`)
	typemeta, err := cv.TypeMetaFromBytes(testZedJSON)
	if err != nil {
		t.Fatal(err)
	}
	objOriginal, err := cv.GetObjectFromBytes(typemeta, testZedJSON)
	if err != nil {
		t.Fatal(err)
	}

	cs := &ConvertSpec{Kinds: []Kind{objOriginal}}
	cs, err = cv.ConvertToOldest(cs, testGroup)
	if err != nil {
		t.Fatalf("failed converting to oldest: %v", err)
	}

	expectedFoo := &testFoo{A: "A"}
	cv.SetGetDefaultTypeMeta(Kind(expectedFoo))
	if !reflect.DeepEqual(cs.Kinds[0], expectedFoo) {
		t.Fatalf("expected oldest:\n%#v\ngot:\n%#v", expectedFoo, cs.Kinds[0])
	}

	cs, err = cv.ConvertToLatest(cs, testGroup)
	if err != nil {
		t.Fatalf("failed converting to latest: %v", err)
	}
	if !reflect.DeepEqual(cs.Kinds[0], objOriginal) {
		t.Fatalf("expected roundtrip to latest:\n%#v\ngot:\n%#v", expectedFoo, cs.Kinds[0])
	}
}

// testFoo
type testFoo struct {
	metav1.TypeMeta `json:",inline"`
	A               string `json:"a"`
}

func (*testFoo) ConvertUp(cv *Converter, in *ConvertSpec) (*ConvertSpec, error)   { return nil, nil }
func (*testFoo) ConvertDown(cv *Converter, in *ConvertSpec) (*ConvertSpec, error) { return nil, nil }
func (*testFoo) ConvertUpSpec() *ConvertSpec                                      { return &ConvertSpec{} }
func (*testFoo) ConvertDownSpec() *ConvertSpec                                    { return &ConvertSpec{Kinds: []Kind{&testFoo{}}} }
func (*testFoo) Validate() error                                                  { return nil }
func (*testFoo) Default() error                                                   { return nil }
func (x *testFoo) GetTypeMeta() *metav1.TypeMeta                                  { return &x.TypeMeta }
func (*testFoo) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: testGroup + "/v1beta1", Kind: "testFoo"}
}

// testBar
type testBar struct {
	metav1.TypeMeta `json:",inline"`
	A               string `json:"a"`
	B               string `json:"b"`
}

func (*testBar) ConvertUp(cv *Converter, in *ConvertSpec) (*ConvertSpec, error) {
	ink := in.Kinds[0]
	new := &testBar{}
	cv.DeepCopy(new, ink)
	cachedKind := cv.GetFromCache(new)
	if cachedKind != nil {
		cached := cachedKind.(*testBar)
		new.B = cached.B
	}
	return &ConvertSpec{Kinds: []Kind{new}}, nil
}
func (*testBar) ConvertDown(cv *Converter, in *ConvertSpec) (*ConvertSpec, error) {
	ink := in.Kinds[0]
	cv.AddToCache(ink)
	new := &testFoo{}
	cv.DeepCopy(new, ink)
	return &ConvertSpec{Kinds: []Kind{new}}, nil
}
func (*testBar) ConvertUpSpec() *ConvertSpec     { return &ConvertSpec{Kinds: []Kind{&testFoo{}}} }
func (*testBar) ConvertDownSpec() *ConvertSpec   { return &ConvertSpec{Kinds: []Kind{&testBar{}}} }
func (*testBar) Validate() error                 { return nil }
func (*testBar) Default() error                  { return nil }
func (x *testBar) GetTypeMeta() *metav1.TypeMeta { return &x.TypeMeta }
func (*testBar) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: testGroup + "/v1beta2", Kind: "testBar"}
}

// testZed
type testZed struct {
	metav1.TypeMeta `json:",inline"`
	A               string `json:"a"`
	B               string `json:"b"`
	C               string `json:"c"`
}

func (*testZed) ConvertUp(cv *Converter, in *ConvertSpec) (*ConvertSpec, error) {
	ink := in.Kinds[0]
	new := &testZed{}
	cv.DeepCopy(new, ink)
	cachedKind := cv.GetFromCache(new)
	if cachedKind != nil {
		cached := cachedKind.(*testZed)
		new.C = cached.C
	}
	return &ConvertSpec{Kinds: []Kind{new}}, nil
}
func (*testZed) ConvertDown(cv *Converter, in *ConvertSpec) (*ConvertSpec, error) {
	ink := in.Kinds[0]
	cv.AddToCache(ink)
	new := &testBar{}
	cv.DeepCopy(new, ink)
	return &ConvertSpec{Kinds: []Kind{new}}, nil
}
func (*testZed) ConvertUpSpec() *ConvertSpec     { return &ConvertSpec{Kinds: []Kind{&testBar{}}} }
func (*testZed) ConvertDownSpec() *ConvertSpec   { return &ConvertSpec{Kinds: []Kind{&testZed{}}} }
func (*testZed) Validate() error                 { return nil }
func (*testZed) Default() error                  { return nil }
func (x *testZed) GetTypeMeta() *metav1.TypeMeta { return &x.TypeMeta }
func (*testZed) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: testGroup + "/v1beta3", Kind: "testZed"}
}
