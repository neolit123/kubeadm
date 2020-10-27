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

func TestSplitDocuments(t *testing.T) {
	foo := `{ "foo": "Foo" }`
	bar := `{ "bar": "Bar" }`
	multiDoc := foo + "\n---\n" + bar

	cv := NewConverter("testgroup", nil)
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
	versionKinds := []VersionKinds{
		{"v1beta1", []Kind{&testFoo{}}},
		{"v1beta2", []Kind{&testBar{}}},
		{"v1beta3", []Kind{&testZed{}}},
	}
	cv := NewConverter("testgroup", versionKinds)
	cv.SetMarshalFunc(json.Marshal)
	cv.SetUnmarshalFunc(json.Unmarshal)

	testZedJSON := []byte(`{"kind": "testZed", "apiVersion": "testgroup/v1beta3", "a": "A", "b": "B", "c": "C"}`)
	typemeta, err := cv.GetTypeMetaFromBytes(testZedJSON)
	if err != nil {
		t.Fatal(err)
	}
	objOriginal, err := cv.GetObjectFromBytes(typemeta, testZedJSON)
	if err != nil {
		t.Fatal(err)
	}

	obj, err := cv.ConvertToOldest(objOriginal)
	if err != nil {
		t.Fatalf("failed converting to oldest: %v", err)
	}
	expectedFoo := &testFoo{A: "A"}
	cv.SetTypeMeta(Kind(expectedFoo))
	if !reflect.DeepEqual(obj, expectedFoo) {
		t.Fatalf("exected oldest:\n%#v\ngot:\n%#v", expectedFoo, obj)
	}

	obj, err = cv.ConvertToLatest(obj)
	if err != nil {
		t.Fatalf("failed converting to latest: %v", err)
	}
	if !reflect.DeepEqual(obj, objOriginal) {
		t.Fatalf("exected roundtrip to latest:\n%#v\ngot:\n%#v", expectedFoo, obj)
	}
}

// testFoo
type testFoo struct {
	metav1.TypeMeta `json:",inline"`
	A               string `json:"a"`
}

func (*testFoo) Version() string                                  { return "v1beta1" }
func (*testFoo) Name() string                                     { return "testFoo" }
func (*testFoo) ConvertUp(cv *Converter, in Kind) (Kind, error)   { return nil, nil }
func (*testFoo) ConvertDown(cv *Converter, in Kind) (Kind, error) { return nil, nil }
func (*testFoo) ConvertUpName() string                            { return "" }
func (*testFoo) Validate() error                                  { return nil }
func (*testFoo) Default() error                                   { return nil }
func (x *testFoo) GetTypeMeta() *metav1.TypeMeta                  { return &x.TypeMeta }

// testBar
type testBar struct {
	metav1.TypeMeta `json:",inline"`
	A               string `json:"a"`
	B               string `json:"b"`
}

func (*testBar) Version() string { return "v1beta2" }
func (*testBar) Name() string    { return "testBar" }
func (*testBar) ConvertUp(cv *Converter, in Kind) (Kind, error) {
	new := &testBar{}
	cv.DeepCopy(new, in)
	cachedKind := cv.GetFromCache(new)
	if cachedKind != nil {
		cached := cachedKind.(*testBar)
		new.B = cached.B
	}
	return new, nil
}
func (*testBar) ConvertDown(cv *Converter, in Kind) (Kind, error) {
	cv.AddToCache(in)
	new := &testFoo{}
	cv.DeepCopy(new, in)
	return new, nil
}
func (*testBar) ConvertUpName() string           { return (&testFoo{}).Name() }
func (*testBar) Validate() error                 { return nil }
func (*testBar) Default() error                  { return nil }
func (x *testBar) GetTypeMeta() *metav1.TypeMeta { return &x.TypeMeta }

// testZed
type testZed struct {
	metav1.TypeMeta `json:",inline"`
	A               string `json:"a"`
	B               string `json:"b"`
	C               string `json:"c"`
}

func (*testZed) Version() string { return "v1beta3" }
func (*testZed) Name() string    { return "testZed" }
func (*testZed) ConvertUp(cv *Converter, in Kind) (Kind, error) {
	new := &testZed{}
	cv.DeepCopy(new, in)
	cachedKind := cv.GetFromCache(new)
	if cachedKind != nil {
		cached := cachedKind.(*testZed)
		new.C = cached.C
	}
	return new, nil
}
func (*testZed) ConvertDown(cv *Converter, in Kind) (Kind, error) {
	cv.AddToCache(in)
	new := &testBar{}
	cv.DeepCopy(new, in)
	return new, nil
}
func (*testZed) ConvertUpName() string           { return (&testBar{}).Name() }
func (*testZed) Validate() error                 { return nil }
func (*testZed) Default() error                  { return nil }
func (x *testZed) GetTypeMeta() *metav1.TypeMeta { return &x.TypeMeta }
