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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testGroup0 = "testgroup0"
	testGroup1 = "testgroup1"
	testGroup2 = "testgroup2"
)

var (
	testGroups = []Group{
		{
			Name: testGroup0,
			Versions: []VersionKinds{
				{"v1", []Kind{&testQue{}}},
			},
		},
		{
			Name: testGroup1,
			Versions: []VersionKinds{
				{"v1beta1", []Kind{&testFoo{}}},
				{"v1beta2", []Kind{&testBar1{}, &testBar2{}}},
				{"v1beta3", []Kind{&testZed{}}},
			},
		},
		{
			Name: testGroup2,
			Versions: []VersionKinds{
				{"v1", []Kind{&testBaz{}}},
			},
		},
	}
	testFooJSON = []byte(`{"kind": "testFoo", "apiVersion": "testgroup1/v1beta1", "A": "A"}`)
	testZedJSON = []byte(`{"kind": "testZed", "apiVersion": "testgroup1/v1beta3", "a": "A", "b": "B", "c": "C"}`)
)

func TestConvertBetweenGroups(t *testing.T) {
	cv := NewConverter().WithGroups(testGroups)

	// convert from older to newer group
	typemeta, err := cv.TypeMetaFromBytes(testZedJSON)
	if err != nil {
		t.Fatal(err)
	}
	obj, err := cv.GetObjectFromBytes(typemeta, testZedJSON)
	if err != nil {
		t.Fatal(err)
	}
	cs := NewKindSpec().WithKinds(obj)
	cs, err = cv.ConvertTo(cs, testGroup2, "v1")
	if err != nil {
		t.Fatalf("failed converting to a newer group: %v", err)
	}
	result, err := cv.Marshal(cs.Kinds[0])
	if err != nil {
		t.Fatal(err)
	}
	expectedResult := `{"kind":"testBaz","apiVersion":"testgroup2/v1","a":"A","b":"B","c":"C","d":""}`
	if expectedResult != string(result) {
		t.Fatalf("expected result:\n%s\ngot:\n%s", expectedResult, result)
	}

	// convert from newer to older group
	typemeta, err = cv.TypeMetaFromBytes(testFooJSON)
	if err != nil {
		t.Fatal(err)
	}
	obj, err = cv.GetObjectFromBytes(typemeta, testFooJSON)
	if err != nil {
		t.Fatal(err)
	}
	cs = NewKindSpec().WithKinds(obj)
	cs, err = cv.ConvertTo(cs, testGroup0, "v1")
	if err != nil {
		t.Fatalf("failed converting to a older group: %v", err)
	}
	result, err = cv.Marshal(cs.Kinds[0])
	if err != nil {
		t.Fatal(err)
	}
	expectedResult = `{"kind":"testQue","apiVersion":"testgroup0/v1","m":"A"}`
	if expectedResult != string(result) {
		t.Fatalf("expected result:\n%s\ngot\n:%s", expectedResult, result)
	}
}

func TestConvert(t *testing.T) {
	cv := NewConverter().WithGroups(testGroups)

	typemeta, err := cv.TypeMetaFromBytes(testZedJSON)
	if err != nil {
		t.Fatal(err)
	}
	objOriginal, err := cv.GetObjectFromBytes(typemeta, testZedJSON)
	if err != nil {
		t.Fatal(err)
	}

	cs := NewKindSpec().WithKinds(objOriginal)
	cs, err = cv.ConvertToOldest(cs, testGroup1)
	if err != nil {
		t.Fatalf("failed converting to oldest: %v", err)
	}

	expectedFoo := &testFoo{A: "A"}
	SetDefaultTypeMeta(Kind(expectedFoo))
	if !reflect.DeepEqual(cs.Kinds[0], expectedFoo) {
		t.Fatalf("expected oldest:\n%#v\ngot:\n%#v", expectedFoo, cs.Kinds[0])
	}

	cs, err = cv.ConvertToLatest(cs, testGroup1)
	if err != nil {
		t.Fatalf("failed converting to latest: %v", err)
	}
	if !reflect.DeepEqual(cs.Kinds[0], objOriginal) {
		t.Fatalf("expected roundtrip back to latest:\n%#v\ngot:\n%#v", objOriginal, cs.Kinds[0])
	}
}

// testQue
type testQue struct {
	metav1.TypeMeta `json:",inline"`
	M               string `json:"m"`
}

func (*testQue) ConvertUp(cv *Converter, in *KindSpec) (*KindSpec, error) {
	return in, nil
}
func (*testQue) ConvertDown(cv *Converter, in *KindSpec) (*KindSpec, error) {
	return in, nil
}
func (*testQue) ConvertUpSpec() *KindSpec        { return NewKindSpec().WithKinds(&testQue{}) }
func (*testQue) ConvertDownSpec() *KindSpec      { return NewKindSpec().WithKinds(&testQue{}) }
func (*testQue) Validate() error                 { return nil }
func (*testQue) Default() error                  { return nil }
func (x *testQue) GetTypeMeta() *metav1.TypeMeta { return &x.TypeMeta }
func (*testQue) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: testGroup0 + "/v1", Kind: "testQue"}
}

// testFoo
type testFoo struct {
	metav1.TypeMeta `json:",inline"`
	A               string `json:"a"`
}

func (*testFoo) ConvertUp(cv *Converter, in *KindSpec) (*KindSpec, error) {
	que := in.Kinds[0]
	foo := &testFoo{}
	SetDefaultTypeMeta(foo)
	foo.A = que.(*testQue).M
	return NewKindSpec().WithKinds(foo), nil
}
func (*testFoo) ConvertDown(cv *Converter, in *KindSpec) (*KindSpec, error) {
	foo := in.Kinds[0]
	que := &testQue{}
	SetDefaultTypeMeta(que)
	que.M = foo.(*testFoo).A
	return NewKindSpec().WithKinds(que), nil
}
func (*testFoo) ConvertUpSpec() *KindSpec        { return NewKindSpec().WithKinds(&testQue{}) }
func (*testFoo) ConvertDownSpec() *KindSpec      { return NewKindSpec().WithKinds(&testFoo{}) }
func (*testFoo) Validate() error                 { return nil }
func (*testFoo) Default() error                  { return nil }
func (x *testFoo) GetTypeMeta() *metav1.TypeMeta { return &x.TypeMeta }
func (*testFoo) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: testGroup1 + "/v1beta1", Kind: "testFoo"}
}

// testBar1
type testBar1 struct {
	metav1.TypeMeta `json:",inline"`
	A               string `json:"a"`
}

func (*testBar1) ConvertUp(cv *Converter, in *KindSpec) (*KindSpec, error) {
	foo := in.Kinds[0]
	bar1 := &testBar1{}
	bar2 := &testBar2{}
	DeepCopy(bar1, foo)
	cachedKind := cv.GetFromCache(bar2)
	if cachedKind != nil {
		cached := cachedKind.(*testBar2)
		bar2.B = cached.B
	}
	return NewKindSpec().WithKinds(bar1, bar2), nil
}
func (*testBar1) ConvertDown(cv *Converter, in *KindSpec) (*KindSpec, error) {
	bar1 := in.Kinds[0]
	bar2 := in.Kinds[1]
	cv.AddToCache(bar2)
	foo := &testFoo{}
	DeepCopy(foo, bar1)
	return NewKindSpec().WithKinds(foo), nil
}
func (*testBar1) ConvertUpSpec() *KindSpec { return NewKindSpec().WithKinds(&testFoo{}) }
func (*testBar1) ConvertDownSpec() *KindSpec {
	return NewKindSpec().WithKinds(&testBar1{}, &testBar2{})
}
func (*testBar1) Validate() error                 { return nil }
func (*testBar1) Default() error                  { return nil }
func (x *testBar1) GetTypeMeta() *metav1.TypeMeta { return &x.TypeMeta }
func (*testBar1) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: testGroup1 + "/v1beta2", Kind: "testBar1"}
}

// testBar2
type testBar2 struct {
	metav1.TypeMeta `json:",inline"`
	B               string `json:"b"`
}

func (*testBar2) ConvertUp(cv *Converter, in *KindSpec) (*KindSpec, error)   { return in, nil }
func (*testBar2) ConvertDown(cv *Converter, in *KindSpec) (*KindSpec, error) { return in, nil }
func (*testBar2) ConvertUpSpec() *KindSpec                                   { return NewKindSpec().WithKinds(&testFoo{}) }
func (*testBar2) ConvertDownSpec() *KindSpec                                 { return NewKindSpec().WithKinds(&testBar2{}) }
func (*testBar2) Validate() error                                            { return nil }
func (*testBar2) Default() error                                             { return nil }
func (x *testBar2) GetTypeMeta() *metav1.TypeMeta                            { return &x.TypeMeta }
func (*testBar2) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: testGroup1 + "/v1beta2", Kind: "testBar2"}
}

// testZed
type testZed struct {
	metav1.TypeMeta `json:",inline"`
	A               string `json:"a"`
	B               string `json:"b"`
	C               string `json:"c"`
}

func (*testZed) ConvertUp(cv *Converter, in *KindSpec) (*KindSpec, error) {
	bar1 := in.Kinds[0]
	bar2 := in.Kinds[1]
	zed := &testZed{}
	DeepCopy(zed, bar1)
	DeepCopy(zed, bar2)
	cachedKind := cv.GetFromCache(zed)
	if cachedKind != nil {
		cached := cachedKind.(*testZed)
		zed.C = cached.C
	}
	return NewKindSpec().WithKinds(zed), nil
}
func (*testZed) ConvertDown(cv *Converter, in *KindSpec) (*KindSpec, error) {
	zed := in.Kinds[0]
	cv.AddToCache(zed)
	bar1 := &testBar1{}
	bar2 := &testBar2{}
	DeepCopy(bar1, zed)
	DeepCopy(bar2, zed)
	return NewKindSpec().WithKinds(bar1, bar2), nil
}
func (*testZed) ConvertUpSpec() *KindSpec        { return NewKindSpec().WithKinds(&testBar1{}, &testBar2{}) }
func (*testZed) ConvertDownSpec() *KindSpec      { return NewKindSpec().WithKinds(&testZed{}) }
func (*testZed) Validate() error                 { return nil }
func (*testZed) Default() error                  { return nil }
func (x *testZed) GetTypeMeta() *metav1.TypeMeta { return &x.TypeMeta }
func (*testZed) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: testGroup1 + "/v1beta3", Kind: "testZed"}
}

// testBaz
type testBaz struct {
	metav1.TypeMeta `json:",inline"`
	A               string `json:"a"`
	B               string `json:"b"`
	C               string `json:"c"`
	D               string `json:"d"`
}

func (*testBaz) ConvertUp(cv *Converter, in *KindSpec) (*KindSpec, error) {
	zed := in.Kinds[0]
	baz := &testBaz{}
	DeepCopy(baz, zed)
	cachedKind := cv.GetFromCache(baz)
	if cachedKind != nil {
		cached := cachedKind.(*testBaz)
		baz.D = cached.D
	}
	return NewKindSpec().WithKinds(baz), nil
}
func (*testBaz) ConvertDown(cv *Converter, in *KindSpec) (*KindSpec, error) {
	baz := in.Kinds[0]
	cv.AddToCache(baz)
	zed := &testZed{}
	DeepCopy(zed, baz)
	return NewKindSpec().WithKinds(zed), nil
}
func (*testBaz) ConvertUpSpec() *KindSpec        { return NewKindSpec().WithKinds(&testZed{}) }
func (*testBaz) ConvertDownSpec() *KindSpec      { return NewKindSpec().WithKinds(&testBaz{}) }
func (*testBaz) Validate() error                 { return nil }
func (*testBaz) Default() error                  { return nil }
func (x *testBaz) GetTypeMeta() *metav1.TypeMeta { return &x.TypeMeta }
func (*testBaz) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: testGroup2 + "/v1", Kind: "testBaz"}
}
