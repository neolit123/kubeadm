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
func (*testQue) ConvertUpSpec() *KindSpec   { return NewKindSpec().WithKinds(&testQue{}) }
func (*testQue) ConvertDownSpec() *KindSpec { return NewKindSpec().WithKinds(&testQue{}) }
func (*testQue) Validate() error            { return nil }
func (*testQue) Default() error             { return nil }
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
func (*testFoo) ConvertUpSpec() *KindSpec   { return NewKindSpec().WithKinds(&testQue{}) }
func (*testFoo) ConvertDownSpec() *KindSpec { return NewKindSpec().WithKinds(&testFoo{}) }
func (*testFoo) Validate() error            { return nil }
func (*testFoo) Default() error             { return nil }
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
func (*testBar1) Validate() error { return nil }
func (*testBar1) Default() error  { return nil }
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
func (*testZed) ConvertUpSpec() *KindSpec   { return NewKindSpec().WithKinds(&testBar1{}, &testBar2{}) }
func (*testZed) ConvertDownSpec() *KindSpec { return NewKindSpec().WithKinds(&testZed{}) }
func (*testZed) Validate() error            { return nil }
func (*testZed) Default() error             { return nil }
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
func (*testBaz) ConvertUpSpec() *KindSpec   { return NewKindSpec().WithKinds(&testZed{}) }
func (*testBaz) ConvertDownSpec() *KindSpec { return NewKindSpec().WithKinds(&testBaz{}) }
func (*testBaz) Validate() error            { return nil }
func (*testBaz) Default() error             { return nil }
func (*testBaz) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: testGroup2 + "/v1", Kind: "testBaz"}
}

func TestDeleteMetadata(t *testing.T) {
	in := []byte(`{"foo":"a","bar":"b","metadata":{}}`)
	cv := NewConverter()
	out, err := cv.DeleteMetadata(in)
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte(`{"bar":"b","foo":"a"}`)
	if !bytes.Equal(expected, out) {
		t.Fatalf("expected:\n%v\ngot:\n%v", expected, out)
	}
}

func TestGetAnnotations(t *testing.T) {
	testCases := []struct {
		name                string
		input               []byte
		expectedAnnotations map[string]string
		expectedError       bool
	}{
		{
			name:                "valid: return annotations",
			input:               []byte(`{"foo":"bar","metadata":{"foo":"bar","annotations":{"a":"a","b":"b"}}}`),
			expectedAnnotations: map[string]string{"a": "a", "b": "b"},
			expectedError:       false,
		},
		{
			name:                "valid: return empty annotations",
			input:               []byte(`{"foo":"bar","metadata":{"annotations":{}}}`),
			expectedAnnotations: map[string]string{},
			expectedError:       false,
		},
		{
			name:          "invalid: annotation value is non-string",
			input:         []byte(`{"foo":"bar","metadata":{"annotations":{"a":1}}}`),
			expectedError: true,
		},
		{
			name:          "invalid: null metadata",
			input:         []byte(`{"foo":"bar","metadata":null}`),
			expectedError: true,
		},
		{
			name:          "invalid: null annotations",
			input:         []byte(`{"foo":"bar","metadata":{"annotations":null}}`),
			expectedError: true,
		},
		{
			name:          "invalid: missing metadata",
			input:         []byte(`{"foo":"bar"}`),
			expectedError: true,
		},
		{
			name:          "invalid: missing annotations",
			input:         []byte(`{"foo":"bar","metadata":{}}`),
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cv := NewConverter()
			annotations, err := cv.GetAnnotations(tc.input)
			if (err != nil) != tc.expectedError {
				t.Fatalf("expected error %v, got %v, error: %v", tc.expectedError, err != nil, err)
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(tc.expectedAnnotations, annotations) {
				t.Fatalf("expected annotations:\n%+v\ngot:\n%+v", tc.expectedAnnotations, annotations)
			}
		})
	}
}

func TestSetAnnotations(t *testing.T) {
	testCases := []struct {
		name           string
		input          []byte
		annotations    map[string]string
		expectedOutput []byte
		expectedError  bool
	}{
		{
			name:           "valid: set annotations",
			input:          []byte(`{"foo":"bar"}`),
			annotations:    map[string]string{"a": "a", "b": "b"},
			expectedOutput: []byte(`{"foo":"bar","metadata":{"annotations":{"a":"a","b":"b"}}}`),
			expectedError:  false,
		},
		{
			name:           "valid: set null annotations",
			input:          []byte(`{"foo":"bar"}`),
			annotations:    nil,
			expectedOutput: []byte(`{"foo":"bar","metadata":{"annotations":null}}`),
			expectedError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cv := NewConverter()
			output, err := cv.SetAnnotations(tc.input, tc.annotations)
			if (err != nil) != tc.expectedError {
				t.Fatalf("expected error %v, got %v, error: %v", tc.expectedError, err != nil, err)
			}
			if err != nil {
				return
			}
			if !bytes.Equal(tc.expectedOutput, output) {
				t.Fatalf("expected output:\n%s\ngot:\n%s", tc.expectedOutput, output)
			}
		})
	}
}

func TestAddCacheToAnnotations(t *testing.T) {
	testCases := []struct {
		name           string
		annotations    map[string]string
		kinds          []Kind
		expectedOutput map[string]string
		expectedError  bool
	}{
		{
			name:        "valid: add kind as annotation",
			annotations: map[string]string{"a": "a", "b": "b"},
			kinds:       []Kind{&testFoo{A: "foo"}},
			expectedOutput: map[string]string{"a": "a", "b": "b",
				ConverterCacheAnnotation + ".testgroup1/v1beta1.testFoo": `{"kind":"testFoo","apiVersion":"testgroup1/v1beta1","a":"foo"}`},
			expectedError: false,
		},
		{
			name:           "valid: empty cache",
			annotations:    map[string]string{"a": "a", "b": "b"},
			kinds:          []Kind{},
			expectedOutput: map[string]string{"a": "a", "b": "b"},
			expectedError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cv := NewConverter()
			for _, k := range tc.kinds {
				cv.AddToCache(k)
			}
			err := cv.AddCacheToAnnotations(tc.annotations)
			if (err != nil) != tc.expectedError {
				t.Fatalf("expected error %v, got %v, error: %v", tc.expectedError, err != nil, err)
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(tc.expectedOutput, tc.annotations) {
				t.Fatalf("expected output:\n%#v\ngot:\n%#v", tc.expectedOutput, tc.annotations)
			}
		})
	}
}

func TestAddAnnotationsToCache(t *testing.T) {
	testCases := []struct {
		name          string
		annotations   map[string]string
		expectedCache map[string]Kind
		expectedError bool
	}{
		{
			name: "valid: add kind with typemeta from annotation to cache",
			annotations: map[string]string{"a": "a", "b": "b",
				ConverterCacheAnnotation + ".testgroup1/v1beta1.testFoo": `{"kind":"testFoo","apiVersion":"testgroup1/v1beta1","a":"foo"}`},
			expectedCache: map[string]Kind{"testgroup1/v1beta1.testFoo": &testFoo{
				A: "foo", TypeMeta: metav1.TypeMeta{APIVersion: "testgroup1/v1beta1", Kind: "testFoo"}}},
			expectedError: false,
		},
		{
			name: "valid: add kind without typemeta from annotation to cache",
			annotations: map[string]string{"a": "a", "b": "b",
				ConverterCacheAnnotation + ".testgroup1/v1beta1.testFoo": `{"a":"foo"}`},
			expectedCache: map[string]Kind{"testgroup1/v1beta1.testFoo": &testFoo{
				A: "foo", TypeMeta: metav1.TypeMeta{APIVersion: "testgroup1/v1beta1", Kind: "testFoo"}}},
			expectedError: false,
		},
		{
			name: "invalid: unknown kind with typemeta stored in annotation",
			annotations: map[string]string{"a": "a", "b": "b",
				ConverterCacheAnnotation + ".testgroup1/v1beta1.testFoo": `{"kind":"unknown","apiVersion":"testgroup1/v1beta1","a":"foo"}`},
			expectedError: true,
		},
		{
			name: "invalid: typemeta does not match annotation suffix",
			annotations: map[string]string{"a": "a", "b": "b",
				ConverterCacheAnnotation + ".testgroup1/v1beta1.unknown": `{"kind":"testFoo","apiVersion":"testgroup1/v1beta1","a":"foo"}`},
			expectedError: true,
		},
		{
			name: "invalid: missing typemeta and cannot parse annotation key",
			annotations: map[string]string{"a": "a", "b": "b",
				ConverterCacheAnnotation + ".foo": `{"a":"foo"}`},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cv := NewConverter().WithGroups(testGroups)
			err := cv.AddAnnotationsToCache(tc.annotations)
			if (err != nil) != tc.expectedError {
				t.Fatalf("expected error %v, got %v, error: %v", tc.expectedError, err != nil, err)
			}
			if err != nil {
				return
			}
			for k, v := range tc.expectedCache {
				valCache, ok := cv.cache[k]
				if !ok {
					t.Fatalf("missing key %q in expected cache", k)
				}
				if !reflect.DeepEqual(v, valCache) {
					t.Fatalf("expected output:\n%#v\ngot:\n%#v", v, valCache)
				}
			}
		})
	}
}
