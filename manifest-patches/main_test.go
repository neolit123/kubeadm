package main

import (
	"reflect"
	"testing"
)

func TestMatchComponentName(t *testing.T) {
	tests := []struct {
		input             string
		expectedComponent string
		expectedError     bool
	}{
		{
			input:             "etcdzzz",
			expectedComponent: "etcd",
		},
		{
			input:             "kube-apiserverzzz",
			expectedComponent: "kube-apiserver",
		},
		{
			input:             "kube-controller-managerzzz",
			expectedComponent: "kube-controller-manager",
		},
		{
			input:             "kube-schedulerzzz",
			expectedComponent: "kube-scheduler",
		},
		{
			input:         "foo",
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			c, err := matchComponentName(tc.input)
			if (err != nil) != tc.expectedError {
				t.Errorf("expected error: %v, got: %v, error: %v", tc.expectedError, err != nil, err)
			}
			if c != tc.expectedComponent {
				t.Fatalf("expected component:\n%#v\ngot:\n%#v\n", tc.expectedComponent, c)
			}
		})
	}
}

func TestParsePatch(t *testing.T) {
	tests := []struct {
		fileName         string
		expectedError    bool
		expectedPatchSet *patchSet
	}{
		{
			fileName:         "etcd.yaml",
			expectedPatchSet: &patchSet{componentName: "etcd", patchType: patchTypes[""]},
		},
		{
			fileName:         "kube-apiserver.yaml",
			expectedPatchSet: &patchSet{componentName: "kube-apiserver", patchType: patchTypes[""]},
		},
		{
			fileName:         "kube-controller-manager.yaml",
			expectedPatchSet: &patchSet{componentName: "kube-controller-manager", patchType: patchTypes[""]},
		},
		{
			fileName:         "kube-scheduler.yaml",
			expectedPatchSet: &patchSet{componentName: "kube-scheduler", patchType: patchTypes[""]},
		},
		{
			fileName:         "etcd+strategic",
			expectedPatchSet: &patchSet{componentName: "etcd", patchType: patchTypes["strategic"]},
		},
		{
			fileName:         "etcd+json",
			expectedPatchSet: &patchSet{componentName: "etcd", patchType: patchTypes["json"]},
		},
		{
			fileName:         "etcd+merge.yaml",
			expectedPatchSet: &patchSet{componentName: "etcd", patchType: patchTypes["merge"]},
		},
		{
			fileName:         "etcd0+merge.yaml",
			expectedPatchSet: &patchSet{componentName: "etcd", patchType: patchTypes["merge"]},
		},
		{
			fileName:         "etcd0.yaml",
			expectedPatchSet: &patchSet{componentName: "etcd", patchType: patchTypes[""]},
		},
		{
			fileName:         "etcd0",
			expectedPatchSet: &patchSet{componentName: "etcd", patchType: patchTypes[""]},
		},
		{
			fileName:         "etcd",
			expectedPatchSet: &patchSet{componentName: "etcd", patchType: patchTypes[""]},
		},
		{
			fileName:         "etcd.+",
			expectedPatchSet: &patchSet{componentName: "etcd", patchType: patchTypes[""]},
		},
		{
			fileName:         "etcd.",
			expectedPatchSet: &patchSet{componentName: "etcd", patchType: patchTypes[""]},
		},
		{
			fileName:         "etcd+",
			expectedPatchSet: &patchSet{componentName: "etcd", patchType: patchTypes[""]},
		},
		{
			fileName:      "foo",
			expectedError: true,
		},
		{
			fileName:      "etcd+bar",
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.fileName, func(t *testing.T) {
			p, err := createPatchSet(tc.fileName, nil, nil)
			if (err != nil) != tc.expectedError {
				t.Errorf("expected error: %v, got: %v, error: %v", tc.expectedError, err != nil, err)
			}
			if !reflect.DeepEqual(p, tc.expectedPatchSet) {
				t.Fatalf("expected patch:\n%#v\ngot:\n%#v\n", tc.expectedPatchSet, p)
			}
		})
	}
}
