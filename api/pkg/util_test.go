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

type testKind struct {
	metav1.TypeMeta
	convertUpSpec   *KindSpec
	convertDownSpec *KindSpec
}

func (*testKind) ConvertFrom(*Converter, *KindSpec) (*KindSpec, error)   { return nil, nil }
func (*testKind) ConvertTo(*Converter, *KindSpec) (*KindSpec, error) { return nil, nil }
func (*testKind) Default() error                                       { return nil }
func (*testKind) Validate() error                                      { return nil }
func (t *testKind) GetDefaultTypeMeta() *metav1.TypeMeta               { return &t.TypeMeta }
func (t *testKind) ConvertFromSpec() *KindSpec                           { return t.convertUpSpec }
func (t *testKind) ConvertToSpec() *KindSpec                         { return t.convertDownSpec }

type testKindWithoutTypeMeta struct{}

func (*testKindWithoutTypeMeta) ConvertFrom(*Converter, *KindSpec) (*KindSpec, error) { return nil, nil }
func (*testKindWithoutTypeMeta) ConvertTo(*Converter, *KindSpec) (*KindSpec, error) {
	return nil, nil
}
func (*testKindWithoutTypeMeta) Default() error                         { return nil }
func (*testKindWithoutTypeMeta) Validate() error                        { return nil }
func (t *testKindWithoutTypeMeta) GetDefaultTypeMeta() *metav1.TypeMeta { return nil }
func (t *testKindWithoutTypeMeta) ConvertFromSpec() *KindSpec             { return nil }
func (t *testKindWithoutTypeMeta) ConvertToSpec() *KindSpec           { return nil }

func TestValidateKindSpec(t *testing.T) {
	testCases := []struct {
		name          string
		spec          *KindSpec
		expectedError bool
	}{
		{
			name:          "valid: valid objects in spec",
			expectedError: false,
			spec: NewKindSpec().WithKinds(
				&testKind{TypeMeta: metav1.TypeMeta{APIVersion: "foo/bar", Kind: "k1"}},
				&testKind{TypeMeta: metav1.TypeMeta{APIVersion: "foo/bar", Kind: "k2"}},
			),
		},
		{
			name:          "invalid: object in spec has empty apiVersion",
			expectedError: true,
			spec:          NewKindSpec().WithKinds(&testKind{TypeMeta: metav1.TypeMeta{APIVersion: "", Kind: "bar"}}),
		},
		{
			name:          "invalid: object in spec has empty kind",
			expectedError: true,
			spec:          NewKindSpec().WithKinds(&testKind{TypeMeta: metav1.TypeMeta{APIVersion: "foo/bar", Kind: ""}}),
		},
		{
			name:          "invalid: objects with different versions",
			expectedError: true,
			spec: NewKindSpec().WithKinds(
				&testKind{TypeMeta: metav1.TypeMeta{APIVersion: "foo/bar", Kind: "k1"}},
				&testKind{TypeMeta: metav1.TypeMeta{APIVersion: "foo/baz", Kind: "k2"}},
			),
		},
		{
			name:          "invalid: kind does not embed typemeta",
			expectedError: true,
			spec: NewKindSpec().WithKinds(
				&testKindWithoutTypeMeta{},
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateKindSpec(tc.spec)
			if (err != nil) != tc.expectedError {
				t.Fatalf("expected error %v, got %v, error: %v", tc.expectedError, err != nil, err)
			}
		})
	}
}

func TestValidateGroups(t *testing.T) {
	testCases := []struct {
		name          string
		groups        []Group
		expectedError bool
	}{
		{
			name:          "valid: passes validation",
			expectedError: false,
			groups: []Group{Group{Group: "foo", AddedIn: "v1.11", Versions: []Version{{Version: "bar", AddedIn: "v1.11", Preferred: true, Kinds: []Kind{
				&testKind{
					TypeMeta:        metav1.TypeMeta{APIVersion: "foo/bar", Kind: "k1"},
					convertUpSpec:   &KindSpec{},
					convertDownSpec: &KindSpec{},
				},
			}}}}},
		},
		{
			name:          "invalid: no preferred versions",
			expectedError: true,
			groups: []Group{Group{Group: "foo", AddedIn: "v1.11", Versions: []Version{{Version: "bar", AddedIn: "v1.11", Kinds: []Kind{
				&testKind{
					TypeMeta:        metav1.TypeMeta{APIVersion: "foo/bar", Kind: "k1"},
					convertUpSpec:   &KindSpec{},
					convertDownSpec: &KindSpec{},
				},
			}}}}},
		},
		{
			name:          "invalid: missing group AddedIn",
			expectedError: true,
			groups:        []Group{Group{Group: "foo"}},
		},
		{
			name:          "invalid: missing version AddedIn",
			expectedError: true,
			groups:        []Group{Group{Group: "foo", AddedIn: "v1.15", Versions: []Version{{Version: "bar"}}}},
		},
		{
			name:          "invalid: all versions are deprecated",
			expectedError: true,
			groups:        []Group{Group{Group: "foo", AddedIn: "v1.15", Versions: []Version{{Version: "bar", AddedIn: "v1.15", Preferred: true, Deprecated: true}}}},
		},
		{
			name:          "invalid: unknown group",
			expectedError: true,
			groups: []Group{Group{Group: "foo", Versions: []Version{{Version: "bar", Kinds: []Kind{
				&testKind{
					TypeMeta:        metav1.TypeMeta{APIVersion: "unknown/bar", Kind: "k1"},
					convertUpSpec:   &KindSpec{},
					convertDownSpec: &KindSpec{},
				},
			}}}}},
		},
		{
			name:          "invalid: unknown version",
			expectedError: true,
			groups: []Group{Group{Group: "foo", Versions: []Version{{Version: "bar", Kinds: []Kind{
				&testKind{
					TypeMeta:        metav1.TypeMeta{APIVersion: "foo/unknown", Kind: "k1"},
					convertUpSpec:   &KindSpec{},
					convertDownSpec: &KindSpec{},
				},
			}}}}},
		},
		{
			name:          "invalid: empty groups",
			expectedError: true,
			groups:        []Group{},
		},
		{
			name:          "invalid: empty group name",
			expectedError: true,
			groups:        []Group{Group{Group: "", AddedIn: "v1.15", Versions: []Version{{Version: "foo", AddedIn: "v1.15", Kinds: []Kind{}}}}},
		},
		{
			name:          "invalid: empty version name",
			expectedError: true,
			groups:        []Group{Group{Group: "foo", AddedIn: "v1.15", Versions: []Version{{Version: "", AddedIn: "v1.15", Kinds: []Kind{}}}}},
		},
		{
			name:          "invalid: object does not match parent group",
			expectedError: true,
			groups: []Group{Group{Group: testGroup0, AddedIn: "v1.15", Versions: []Version{{Version: "foo", AddedIn: "v1.15", Kinds: []Kind{
				&testKind{
					TypeMeta: metav1.TypeMeta{APIVersion: testGroup1 + "/foo", Kind: "bar"},
				},
			}}}}},
		},
		{
			name:          "invalid: object does not match parent version",
			expectedError: true,
			groups: []Group{Group{Group: testGroup0, AddedIn: "v1.15", Versions: []Version{{Version: "foo", AddedIn: "v1.15", Kinds: []Kind{
				&testKind{
					TypeMeta: metav1.TypeMeta{APIVersion: testGroup0 + "/bar", Kind: "baz"},
				},
			}}}}},
		},
		{
			name:          "invalid: object has empty kind",
			expectedError: true,
			groups: []Group{Group{Group: testGroup0, AddedIn: "v1.15", Versions: []Version{{Version: "foo", AddedIn: "v1.15", Kinds: []Kind{
				&testKind{
					TypeMeta: metav1.TypeMeta{APIVersion: testGroup0 + "/foo", Kind: ""},
				},
			}}}}},
		},
		{
			name:          "invalid: object does not embed typemeta",
			expectedError: true,
			groups: []Group{Group{Group: testGroup0, AddedIn: "v1.15", Versions: []Version{{Version: "foo", AddedIn: "v1.15", Kinds: []Kind{
				&testKindWithoutTypeMeta{},
			}}}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGroups(tc.groups)
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

	docs, err := SplitDocuments([]byte(multiDoc))
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

func TestGetTypeMeta(t *testing.T) {
	var (
		someNonStruct      = "foo"
		structWithTypeMeta = struct {
			metav1.TypeMeta
			foo string
		}{
			TypeMeta: metav1.TypeMeta{APIVersion: "foo/bar", Kind: "baz"},
			foo:      "foo",
		}
		structWithoutTypeMeta = struct {
			foo string
		}{
			foo: "foo",
		}
		structWithTypeMetaString = struct {
			TypeMeta string
		}{
			TypeMeta: "foo",
		}
	)

	testCases := []struct {
		name             string
		object           interface{}
		expectedTypeMeta *metav1.TypeMeta
		expectedError    bool
	}{
		{
			name:             "valid: matching typemeta pointer",
			object:           &structWithTypeMeta,
			expectedError:    false,
			expectedTypeMeta: &structWithTypeMeta.TypeMeta,
		},
		{
			name:          "invalid: received nil",
			object:        nil,
			expectedError: true,
		},
		{
			name:          "invalid: received non-pointer",
			object:        structWithTypeMeta,
			expectedError: true,
		},
		{
			name:          "invalid: received non-struct object",
			object:        &someNonStruct,
			expectedError: true,
		},
		{
			name:          "invalid: received struct without typemeta",
			object:        &structWithoutTypeMeta,
			expectedError: true,
		},
		{
			name:          "invalid: typemeta field is wrong type",
			object:        &structWithTypeMetaString,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tm, err := getTypeMeta(tc.object)
			if (err != nil) != tc.expectedError {
				t.Fatalf("expected error %v, got %v, error: %v", tc.expectedError, err != nil, err)
			}
			if tm != tc.expectedTypeMeta {
				t.Fatalf("expected typemeta %#v(%p), got %#v(%p)", tc.expectedTypeMeta, tc.expectedTypeMeta, tm, tm)
			}
		})
	}
}

func TestAPIVersionForComponentVersion(t *testing.T) {
	testCases := []struct {
		name            string
		spec            *APIVersionSpec
		expectedVersion *Version
		expectedError   bool
	}{
		{
			name: "valid: found valid version without preferred flag",
			spec: &APIVersionSpec{
				Groups: []Group{{Group: "foo", Versions: []Version{
					{Version: "v1", AddedIn: "v1.14.0", Preferred: false},
					{Version: "v2", AddedIn: "v1.15.0", Preferred: false},
					{Version: "v3", AddedIn: "v1.17.0", Preferred: false},
				}}},
				Group:   "foo",
				CompVer: "v1.16.0",
			},
			expectedVersion: &Version{Version: "v2", AddedIn: "v1.15.0", Preferred: false},
		},
		{
			name: "valid: found valid version with preferred flag",
			spec: &APIVersionSpec{
				Groups: []Group{{Group: "foo", Versions: []Version{
					{Version: "v1", AddedIn: "v1.14.0", Preferred: true},
					{Version: "v2", AddedIn: "v1.15.0", Preferred: false},
					{Version: "v3", AddedIn: "v1.17.0", Preferred: false},
				}}},
				Group:        "foo",
				CompVer:      "v1.16.0",
				UsePreferred: true,
			},
			expectedVersion: &Version{Version: "v1", AddedIn: "v1.14.0", Preferred: true},
		},
		{
			name: "valid: component version is newer use latest",
			spec: &APIVersionSpec{
				Groups: []Group{{Group: "foo", Versions: []Version{
					{Version: "v1", AddedIn: "v1.11.0", Preferred: true},
					{Version: "v2", AddedIn: "v1.12.0", Preferred: false},
					{Version: "v3", AddedIn: "v1.13.0", Preferred: false},
				}}},
				Group:   "foo",
				CompVer: "v1.16.0",
			},
			expectedVersion: &Version{Version: "v3", AddedIn: "v1.13.0", Preferred: false},
		},
		{
			name: "valid: component version is newer use latest preferred",
			spec: &APIVersionSpec{
				Groups: []Group{{Group: "foo", Versions: []Version{
					{Version: "v1", AddedIn: "v1.11.0", Preferred: true},
					{Version: "v2", AddedIn: "v1.12.0", Preferred: false},
					{Version: "v3", AddedIn: "v1.13.0", Preferred: false},
				}}},
				Group:        "foo",
				UsePreferred: true,
				CompVer:      "v1.16.0",
			},
			expectedVersion: &Version{Version: "v1", AddedIn: "v1.11.0", Preferred: true},
		},
		{
			name: "valid: use custom lessEq",
			spec: &APIVersionSpec{
				Groups: []Group{{Group: "foo", Versions: []Version{
					{Version: "v1", AddedIn: "v1.11.0", Preferred: false},
					{Version: "v2", AddedIn: "v1.12.0", Preferred: false},
					{Version: "v3", AddedIn: "v1.13.0", Preferred: false},
				}}},
				Group:   "foo",
				CompVer: "v1.11.0",
				LessEq: func(a string, b string) bool {
					return a == b
				},
			},
			expectedVersion: &Version{Version: "v1", AddedIn: "v1.11.0", Preferred: false},
		},
		{
			name: "valid: passing empty compVer returns latest",
			spec: &APIVersionSpec{
				Groups: []Group{{Group: "foo", Versions: []Version{
					{Version: "v1", AddedIn: "v1.11.0", Preferred: false},
					{Version: "v2", AddedIn: "v1.12.0", Preferred: false},
					{Version: "v3", AddedIn: "v1.13.0", Preferred: false},
				}}},
				Group:   "foo",
				CompVer: "",
			},
			expectedVersion: &Version{Version: "v3", AddedIn: "v1.13.0", Preferred: false},
		},
		{
			name:          "invalid: nil spec",
			expectedError: true,
		},
		{
			name: "invalid: version too old",
			spec: &APIVersionSpec{
				Groups: []Group{{Group: "foo", Versions: []Version{
					{Version: "v1", AddedIn: "v1.11.0", Preferred: true},
					{Version: "v2", AddedIn: "v1.12.0", Preferred: false},
					{Version: "v3", AddedIn: "v1.13.0", Preferred: false},
				}}},
				Group:   "foo",
				CompVer: "v1.10.0",
			},
			expectedError: true,
		},
		{
			name: "invalid: cannot parse compVer",
			spec: &APIVersionSpec{
				Groups:  []Group{{Group: "foo"}},
				Group:   "foo",
				CompVer: "bar",
			},
			expectedError: true,
		},
		{
			name: "invalid: cannot find group",
			spec: &APIVersionSpec{
				Groups: []Group{{Group: "foo"}},
				Group:  "bar",
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ver, err := APIVersionForComponentVersion(tc.spec)
			if (err != nil) != tc.expectedError {
				t.Fatalf("expected error %v, got %v, error: %v", tc.expectedError, err != nil, err)
			}
			if err != nil {
				return
			}
			if !reflect.DeepEqual(tc.expectedVersion, ver) {
				t.Fatalf("expected version %#v, got %#v", tc.expectedVersion, ver)
			}
		})
	}
}
