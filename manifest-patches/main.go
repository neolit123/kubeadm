package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	// "regexp"

	"github.com/pkg/errors"
	// jsonpatch "github.com/evanphx/json-patch/v5"
	// v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	// "k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/yaml"
)

// https://github.com/kubernetes/kubernetes/blob/053d46c3c0a3690b6f9afa1df08b8ef71115b026/staging/src/k8s.io/kubectl/pkg/cmd/patch/patch.go#L298-L330
// https://github.com/kubernetes-sigs/kind/blob/a944589ec78b53fe62b45e8890e45e8d6c078f53/pkg/cluster/internal/patch/resource.go
// https://kubernetes.io/docs/tasks/run-application/update-api-object-kubectl-patch/#use-a-strategic-merge-patch-to-update-a-deployment

// kube-apiserver+strategic.yaml
// kube-apiserver+merge.yaml
// kube-apiserver{0}+json.json

/*

Path to a folder that contains files named "componentname[index][+patchtype][.extension]".
For example "kube-apiserver0+merge.yaml" or just "kube-apiserver.yaml". "index" is a number starting
from zero and determines which patch is applied first. "patchtype" can be one of "strategic", "merge"
or "json" and they match the patch formats supported by kubectl. The default "patchtype" is "strategic".
*/

type patchSet struct {
	componentName string
	patchType     types.PatchType
	patches       [][]byte
}

var patchTypes = map[string]types.PatchType{
	"json":      types.JSONPatchType,
	"merge":     types.MergePatchType,
	"strategic": types.StrategicMergePatchType,
	// Threat an empty value as the default = strategic.
	"": types.StrategicMergePatchType,
}

var components = []string{
	"etcd",
	"kube-apiserver",
	"kube-controller-manager",
	"kube-scheduler",
}

// matchComponentName checks if the given string is prefixed by a known component name.
func matchComponentName(input string) (string, error) {
	var componentName string
	for _, c := range components {
		if strings.HasPrefix(input, c) {
			componentName = c
			break
		}
	}
	if len(componentName) == 0 {
		return "", errors.Errorf("component name must be one of %v from input %q", components, input)
	}
	return componentName, nil
}

// matchPatchType extracts a patch type from a string.
func matchPatchType(input string) (types.PatchType, error) {
	idxDot := strings.LastIndex(input, ".")
	if idxDot == -1 {
		// Shift the dot index into the length of the string.
		idxDot = len(input)
	}

	idxPlus := strings.Index(input, "+")
	if idxPlus == -1 || idxPlus > idxDot {
		// If the + character is missing or if its index is after the dot,
		// just shift it into the dot index.
		idxPlus = idxDot
	} else {
		// Increment the plus index to discard the + character itself.
		idxPlus++
	}

	patchType := input[idxPlus:idxDot]
	pt, ok := patchTypes[patchType]
	if !ok {
		return "", errors.Errorf("unknown patch type %q from input %q", patchType, input)
	}
	return pt, nil
}

// createPatchSet creates a patch object. If componentName is nil it will extract it from the filePath.
// The expected file name format is "componentname[suffix][+patchtype][.extension]".
func createPatchSet(fileName string, componentName *string, data []byte) (*patchSet, error) {
	if componentName == nil {
		// Match the component name from fileName as a prefix.
		c, err := matchComponentName(fileName)
		if err != nil {
			return nil, err
		}
		componentName = &c
	}

	pt, err := matchPatchType(fileName)
	if err != nil {
		return nil, err
	}

	var patches [][]byte
	if len(data) > 0 {
		// Split the patches and convert them to JSON.
		patches = bytes.Split(data, []byte("\n---"))
		for i, patch := range patches {
			patchJSON, err := yaml.YAMLToJSON(patch)
			if err != nil {
				return nil, errors.Wrapf(err, "could not convert patch to JSON:\n%s\n", patch)
			}
			patches[i] = patchJSON
		}
	}

	// Create patchSet.
	return &patchSet{
		componentName: *componentName,
		patchType:     pt,
		patches:       patches,
	}, nil
}

func listFilesForPath(targetPath string) ([]string, error) {
	var files []string
	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "could not list files for path %q", targetPath)
	}
	return files, nil
}

func readPatchesFromPath(patchPath string) ([]*patchSet, error) {
	files, err := listFilesForPath(patchPath)
	if err != nil {
		return nil, err
	}

	patchSets := make([]*patchSet, 0, len(files))
	for _, f := range files {
		// Extract the file name.
		fileName := filepath.Base(f)

		// Match the component name; else print a warning and skip.
		componentName, err := matchComponentName(fileName)
		if err != nil {
			// TODO print warning
			continue
		}

		// Read the patch file.
		data, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read patch from file %q", f)
		}

		// Create a patchSet object.
		patchSet, err := createPatchSet(fileName, &componentName, data)
		if err != nil {
			return nil, err
		}
		patchSets = append(patchSets, patchSet)
	}

	return patchSets, nil
}

func process(file string, patchPath string) error {
	_, err := readPatchesFromPath(patchPath)
	if err != nil {
		return err
	}

	return nil
}

// pt := patchTypes["json"]

// switch pt {
// case types.JSONPatchType:
// case types.MergePatchType:
// case types.StrategicMergePatchType:
// default:
// }

// updated, err := strategicpatch.StrategicMergePatch(dataJSON, patchJSON, v1.Pod{})
// if err != nil {
// 	return err
// }

// patchObj, err := jsonpatch.DecodePatch(patchJSON)
// if err != nil {
// 	return err
// }
// updated, err := patchObj.Apply(dataJSON)
// if err != nil {
// 	return err
// }

// updated, err := jsonpatch.MergePatch(dataJSON, patchJSON)
// if err != nil {
// 	return err
// }

// updatedYAML, err := yaml.JSONToYAML(updated)
// if err != nil {
// 	return err
// }

// fmt.Printf("\n\nupdated:\n%s\n", updatedYAML)
// return nil

/*

go run main.go ~/go/src/k8s.io/kubernetes/_manifests/kube-apiserver.yaml ./patches

*/

func main() {
	if len(os.Args) < 3 {
		fmt.Println("missing args")
		os.Exit(1)
	}
	if err := process(os.Args[1], os.Args[2]); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
		return
	}
}
