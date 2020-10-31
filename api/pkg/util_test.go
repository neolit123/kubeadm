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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type testKind struct {
	convertUpSpec   *KindSpec
	convertDownSpec *KindSpec
	typeMeta        *metav1.TypeMeta
}

func (*testKind) ConvertUp(*Converter, *KindSpec) (*KindSpec, error)   { return nil, nil }
func (*testKind) ConvertDown(*Converter, *KindSpec) (*KindSpec, error) { return nil, nil }
func (*testKind) Default() error                                       { return nil }
func (*testKind) Validate() error                                      { return nil }
func (*testKind) GetTypeMeta() *metav1.TypeMeta                        { return nil }
func (t *testKind) GetDefaultTypeMeta() *metav1.TypeMeta               { return t.typeMeta }
func (t *testKind) ConvertUpSpec() *KindSpec                           { return t.convertUpSpec }
func (t *testKind) ConvertDownSpec() *KindSpec                         { return t.convertDownSpec }

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
				&testKind{nil, nil, &metav1.TypeMeta{APIVersion: "foo/bar", Kind: "k1"}},
				&testKind{nil, nil, &metav1.TypeMeta{APIVersion: "foo/bar", Kind: "k2"}},
			),
		},
		{
			name:          "invalid: object in spec has empty apiVersion",
			expectedError: true,
			spec:          NewKindSpec().WithKinds(&testKind{nil, nil, &metav1.TypeMeta{APIVersion: "", Kind: "bar"}}),
		},
		{
			name:          "invalid: object in spec has empty kind",
			expectedError: true,
			spec:          NewKindSpec().WithKinds(&testKind{nil, nil, &metav1.TypeMeta{APIVersion: "foo/bar", Kind: ""}}),
		},
		{
			name:          "invalid: found objects with different versions",
			expectedError: true,
			spec: NewKindSpec().WithKinds(
				&testKind{nil, nil, &metav1.TypeMeta{APIVersion: "foo/bar", Kind: "k1"}},
				&testKind{nil, nil, &metav1.TypeMeta{APIVersion: "foo/baz", Kind: "k2"}},
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
			groups: []Group{Group{Name: "foo", Versions: []VersionKinds{{Version: "bar", Kinds: []Kind{
				&testKind{
					&KindSpec{}, &KindSpec{}, &metav1.TypeMeta{APIVersion: "foo/bar", Kind: "k1"},
				},
			}}}}},
		},
		{
			name:          "invalid: unknown group",
			expectedError: true,
			groups: []Group{Group{Name: "foo", Versions: []VersionKinds{{Version: "bar", Kinds: []Kind{
				&testKind{
					&KindSpec{}, &KindSpec{}, &metav1.TypeMeta{APIVersion: "unknown/bar", Kind: "k1"},
				},
			}}}}},
		},
		{
			name:          "invalid: unknown version",
			expectedError: true,
			groups: []Group{Group{Name: "foo", Versions: []VersionKinds{{Version: "bar", Kinds: []Kind{
				&testKind{
					&KindSpec{}, &KindSpec{}, &metav1.TypeMeta{APIVersion: "foo/unknown", Kind: "k1"},
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
			groups:        []Group{Group{Name: "", Versions: []VersionKinds{{Version: "foo", Kinds: []Kind{}}}}},
		},
		{
			name:          "invalid: empty version name",
			expectedError: true,
			groups:        []Group{Group{Name: "foo", Versions: []VersionKinds{{Version: "", Kinds: []Kind{}}}}},
		},
		{
			name:          "invalid: object does not match parent group",
			expectedError: true,
			groups: []Group{Group{Name: testGroup0, Versions: []VersionKinds{{Version: "foo", Kinds: []Kind{&testKind{
				nil, nil, &metav1.TypeMeta{APIVersion: testGroup1 + "/foo", Kind: "bar"},
			}}}}}},
		},
		{
			name:          "invalid: object does not match parent version",
			expectedError: true,
			groups: []Group{Group{Name: testGroup0, Versions: []VersionKinds{{Version: "foo", Kinds: []Kind{&testKind{
				nil, nil, &metav1.TypeMeta{APIVersion: testGroup0 + "/bar", Kind: "baz"},
			}}}}}},
		},
		{
			name:          "invalid: object has empty kind",
			expectedError: true,
			groups: []Group{Group{Name: testGroup0, Versions: []VersionKinds{{Version: "foo", Kinds: []Kind{
				&testKind{
					nil, nil, &metav1.TypeMeta{APIVersion: testGroup0 + "/foo", Kind: ""},
				},
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
