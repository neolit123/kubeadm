package main

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/types"
)

func TestMatchComponentName(t *testing.T) {
	tests := []struct {
		fileName          string
		expectedComponent string
		expectedError     bool
	}{
		{
			fileName:          "etcdzzz",
			expectedComponent: "etcd",
		},
		{
			fileName:          "kube-apiserverzzz",
			expectedComponent: "kube-apiserver",
		},
		{
			fileName:          "kube-controller-managerzzz",
			expectedComponent: "kube-controller-manager",
		},
		{
			fileName:          "kube-schedulerzzz",
			expectedComponent: "kube-scheduler",
		},
		{
			fileName:      "foo",
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.fileName, func(t *testing.T) {
			c, err := getComponentNameFromFilename(tc.fileName)
			if (err != nil) != tc.expectedError {
				t.Errorf("expected error: %v, got: %v, error: %v", tc.expectedError, err != nil, err)
			}
			if c != tc.expectedComponent {
				t.Fatalf("expected component:\n%#v\ngot:\n%#v\n", tc.expectedComponent, c)
			}
		})
	}
}

func TestMatchPatchType(t *testing.T) {
	tests := []struct {
		fileName          string
		expectedPatchType types.PatchType
		expectedError     bool
	}{
		{
			fileName:          "foo+strategic.yaml",
			expectedPatchType: types.StrategicMergePatchType,
		},
		{
			fileName:          "foo+json.yaml",
			expectedPatchType: types.JSONPatchType,
		},
		{
			fileName:          "foo+merge.yaml",
			expectedPatchType: types.MergePatchType,
		},
		{
			fileName:          "strategic",
			expectedPatchType: types.StrategicMergePatchType,
		},
		{
			fileName:      "foo+bar.yaml",
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.fileName, func(t *testing.T) {
			pt, err := getPatchTypeFromFilename(tc.fileName)
			if (err != nil) != tc.expectedError {
				t.Fatalf("expected error: %v, got: %v, error: %v", tc.expectedError, err != nil, err)
			}
			if pt != tc.expectedPatchType {
				t.Fatalf("expected patchType: %s, got: %s", tc.expectedPatchType, pt)
			}
		})
	}
}

func TestCreatePatchSet(t *testing.T) {
	tests := []struct {
		name             string
		componentName    string
		patchType        types.PatchType
		expectedPatchSet *patchSet
		data             string
	}{
		{
			name:          "valid: YAML patches are separated and converted to JSON",
			componentName: "etcd",
			patchType:     types.StrategicMergePatchType,
			data:          "foo: bar\n---foo: baz\n",
			expectedPatchSet: &patchSet{
				componentName: "etcd",
				patchType:     types.StrategicMergePatchType,
				patches:       []string{"{\"foo\":\"bar\"}", "{\"foo\":\"baz\"}"},
			},
		},
		{
			name:          "valid: JSON patches are separated",
			componentName: "etcd",
			patchType:     types.StrategicMergePatchType,
			data:          "{\"foo\":\"bar\"}\n---{\"foo\":\"baz\"}",
			expectedPatchSet: &patchSet{
				componentName: "etcd",
				patchType:     types.StrategicMergePatchType,
				patches:       []string{"{\"foo\":\"bar\"}", "{\"foo\":\"baz\"}"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, _ := createPatchSet(tc.componentName, tc.patchType, tc.data)
			if !reflect.DeepEqual(p, tc.expectedPatchSet) {
				t.Fatalf("expected patch:\n%#v\ngot:\n%#v\n", tc.expectedPatchSet, p)
			}
		})
	}
}
