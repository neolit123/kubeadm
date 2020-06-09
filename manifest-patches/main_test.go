package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

func TestFilerPatchFilesForPath(t *testing.T) {
	tests := []struct {
		name                 string
		filesToWrite         []string
		expectedPatchFiles   []*patchFile
		expectedIgnoredFiles []string
	}{
		{
			name:         "valid: patch files are sorted and non-patch files are ignored",
			filesToWrite: []string{"kube-scheduler.json", "kube-apiserver.yaml", "etcd.yaml", "foo", "bar.json"},
			expectedPatchFiles: []*patchFile{
				{
					path:          "etcd.yaml",
					componentName: "etcd",
				},
				{
					path:          "kube-apiserver.yaml",
					componentName: "kube-apiserver",
				},
				{
					path:          "kube-scheduler.json",
					componentName: "kube-scheduler",
				},
			},
			expectedIgnoredFiles: []string{"bar.json", "foo"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir, err := ioutil.TempDir("", "list-patch-files")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)
			for _, file := range tc.filesToWrite {
				filePath := filepath.Join(tempDir, file)
				err := ioutil.WriteFile(filePath, []byte{}, 0644)
				if err != nil {
					t.Fatalf("could not write temporary file %q", filePath)
				}
			}
			patchFiles, ignoredFiles, err := filerPatchFilesForPath(tempDir)
			if err != nil {
				t.Fatal(err)
			}

			// Update paths to include temp. dir.
			for i := range tc.expectedPatchFiles {
				tc.expectedPatchFiles[i].path = filepath.Join(tempDir, tc.expectedPatchFiles[i].path)
			}
			for i := range tc.expectedIgnoredFiles {
				tc.expectedIgnoredFiles[i] = filepath.Join(tempDir, tc.expectedIgnoredFiles[i])
			}

			if len(tc.expectedPatchFiles) != len(patchFiles) {
				t.Fatalf("expected patch files:\n%+v\ngot:\n%+v", tc.expectedPatchFiles, patchFiles)
			}
			for i := range tc.expectedPatchFiles {
				if *tc.expectedPatchFiles[i] != *patchFiles[i] {
					t.Fatalf("expected patch file at position %d:\n%v\ngot:\n%v", i, *tc.expectedPatchFiles[i], *patchFiles[i])
				}
			}

			if len(tc.expectedIgnoredFiles) != len(ignoredFiles) {
				t.Fatalf("expected files to be ignored:\n%v\ngot:\n%v", tc.expectedIgnoredFiles, ignoredFiles)
			}
			for i := range tc.expectedIgnoredFiles {
				if tc.expectedIgnoredFiles[i] != ignoredFiles[i] {
					t.Fatalf("expected files to be ignored:\n%v\ngot:\n%v", tc.expectedIgnoredFiles, ignoredFiles)
				}
			}
		})
	}
}
