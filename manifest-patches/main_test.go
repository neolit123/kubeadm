package main

import (
	"bytes"
	"io/ioutil"
	"k8s.io/api/core/v1"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/types"
)

var testKnownTargets = []string{
	"etcd",
	"kube-apiserver",
	"kube-controller-manager",
	"kube-scheduler",
}

const testDirPattern = "patch-files"

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
			c, err := getComponentNameFromFilename(tc.fileName, testKnownTargets)
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
			data:          "foo: bar\n---\nfoo: baz\n",
			expectedPatchSet: &patchSet{
				componentName: "etcd",
				patchType:     types.StrategicMergePatchType,
				patches:       []string{`{"foo":"bar"}`, `{"foo":"baz"}`},
			},
		},
		{
			name:          "valid: JSON patches are separated",
			componentName: "etcd",
			patchType:     types.StrategicMergePatchType,
			data:          `{"foo":"bar"}` + "\n---\n" + `{"foo":"baz"}`,
			expectedPatchSet: &patchSet{
				componentName: "etcd",
				patchType:     types.StrategicMergePatchType,
				patches:       []string{`{"foo":"bar"}`, `{"foo":"baz"}`},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ps, _ := createPatchSet(tc.componentName, tc.patchType, tc.data)
			if !reflect.DeepEqual(ps, tc.expectedPatchSet) {
				t.Fatalf("expected patch set:\n%+v\ngot:\n%+v\n", tc.expectedPatchSet, ps)
			}
		})
	}
}

func TestGetPatchFilesForPath(t *testing.T) {
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
			tempDir, err := ioutil.TempDir("", testDirPattern)
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

			patchFiles, ignoredFiles, err := getPatchFilesForPath(tempDir, testKnownTargets)
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

			if !reflect.DeepEqual(tc.expectedIgnoredFiles, ignoredFiles) {
				t.Fatalf("expected ignored files:\n%+v\ngot:\n%+v", tc.expectedIgnoredFiles, ignoredFiles)
			}
			if !reflect.DeepEqual(tc.expectedPatchFiles, patchFiles) {
				t.Fatalf("expected patch files:\n%+v\ngot:\n%+v", tc.expectedPatchFiles, patchFiles)
			}
		})
	}
}

func TestCreatePatchSetsFromPatchFiles(t *testing.T) {
	tests := []struct {
		name              string
		patchFiles        []*patchFile
		expectedPatchSets []*patchSet
	}{
		{
			name: "valid: patch sets are correctly extracted from files",
			patchFiles: []*patchFile{
				{
					path:          "kube-apisever+merge.yaml",
					componentName: "kube-apisever",
					data:          "foo: bar\n---\nfoo: baz",
				},
			},
			expectedPatchSets: []*patchSet{
				{
					componentName: "kube-apisever",
					patchType:     types.MergePatchType,
					patches: []string{
						`{"foo":"bar"}`,
						`{"foo":"baz"}`,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir, err := ioutil.TempDir("", testDirPattern)
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			for _, file := range tc.patchFiles {
				filePath := filepath.Join(tempDir, file.path)
				err := ioutil.WriteFile(filePath, []byte(file.data), 0644)
				if err != nil {
					t.Fatalf("could not write temporary file %q", filePath)
				}
				(*file).path = filePath
			}

			patchSets, err := createPatchSetsFromPatchFiles(tc.patchFiles)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tc.expectedPatchSets, patchSets) {
				t.Fatalf("expected patch sets:\n%+v\ngot:\n%+v", tc.expectedPatchSets, patchSets)
			}
		})
	}
}

func TestPatchComponentTarget(t *testing.T) {
	type file struct {
		name string
		data string
	}

	tests := []struct {
		name            string
		files           []*file
		componentTarget *ComponentTarget
		expectedData    []byte
		expectedError   bool
	}{
		{
			name: "valid: patch a kube-apiserver target using merge patch; json patch is applied first",
			componentTarget: &ComponentTarget{
				ComponentName:             "kube-apiserver",
				StrategicMergePatchObject: v1.Pod{},
				Data:                      []byte("foo: bar\nbaz: qux\n"),
			},
			expectedData: []byte(`{"baz":"qux","foo":"patched"}`),
			files: []*file{
				{
					name: "kube-apiserver+merge.yaml",
					data: "foo: patched",
				},
				{
					name: "kube-apiserver+json.json",
					data: `[{"op": "replace", "path": "/foo", "value": "zzz"}]`,
				},
			},
		},
		{
			name: "valid: kube-apiserver target is patched with json patch",
			componentTarget: &ComponentTarget{
				ComponentName:             "kube-apiserver",
				StrategicMergePatchObject: v1.Pod{},
				Data:                      []byte("foo: bar\n"),
			},
			expectedData: []byte(`{"foo":"zzz"}`),
			files: []*file{
				{
					name: "kube-apiserver+json.json",
					data: `[{"op": "replace", "path": "/foo", "value": "zzz"}]`,
				},
			},
		},
		{
			name: "valid: kube-apiserver target is patched with strategic merge patch",
			componentTarget: &ComponentTarget{
				ComponentName:             "kube-apiserver",
				StrategicMergePatchObject: v1.Pod{},
				Data:                      []byte("foo: bar\n"),
			},
			expectedData: []byte(`{"foo":"zzz"}`),
			files: []*file{
				{
					name: "kube-apiserver+strategic.json",
					data: `{"foo":"zzz"}`,
				},
			},
		},
		{
			name: "valid: etcd target is not changed because there are no patches for it",
			componentTarget: &ComponentTarget{
				ComponentName:             "etcd",
				StrategicMergePatchObject: v1.Pod{},
				Data:                      []byte("foo: bar\n"),
			},
			expectedData: []byte("foo: bar\n"),
			files: []*file{
				{
					name: "kube-apiserver+merge.yaml",
					data: "foo: patched",
				},
			},
		},
		{
			name: "invalid: cannot patch etcd target due to malformed json patch",
			componentTarget: &ComponentTarget{
				ComponentName:             "etcd",
				StrategicMergePatchObject: v1.Pod{},
				Data:                      []byte("foo: bar\n"),
			},
			files: []*file{
				{
					name: "etcd+json.json",
					data: `{"foo":"zzz"}`,
				},
			},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir, err := ioutil.TempDir("", testDirPattern)
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			for _, file := range tc.files {
				filePath := filepath.Join(tempDir, file.name)
				err := ioutil.WriteFile(filePath, []byte(file.data), 0644)
				if err != nil {
					t.Fatalf("could not write temporary file %q", filePath)
				}
			}

			pm, err := GetPatchManagerFromPath(tempDir, testKnownTargets, nil)
			if err != nil {
				t.Fatal(err)
			}

			err = pm.PatchComponentTarget(tc.componentTarget)
			if (err != nil) != tc.expectedError {
				t.Fatalf("expected error: %v, got: %v, error: %v", tc.expectedError, err != nil, err)
			}
			if err != nil {
				return
			}

			if !bytes.Equal(tc.componentTarget.Data, tc.expectedData) {
				t.Fatalf("expected result:\n%s\ngot:\n%s", tc.expectedData, tc.componentTarget.Data)
			}
		})
	}
}
