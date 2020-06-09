package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
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
	patches       []string
}

type componentPatchTarget struct {
	name                 string
	data                 []byte
	strategicPatchObject interface{}
}

type patchFile struct {
	componentName string
	path          string
}

var patchTypes = map[string]types.PatchType{
	"json":      types.JSONPatchType,
	"merge":     types.MergePatchType,
	"strategic": types.StrategicMergePatchType,
	"":          types.StrategicMergePatchType, // Treat an empty value as the default = strategic.
}

var components = []string{
	"etcd",
	"kube-apiserver",
	"kube-controller-manager",
	"kube-scheduler",
}

func getComponentNameFromFilename(fileName string) (string, error) {
	var componentName string

	for _, c := range components {
		if strings.HasPrefix(fileName, c) {
			componentName = c
			break
		}
	}

	if len(componentName) == 0 {
		return "", errors.Errorf("component name must be one of %v from file name %q", components, fileName)
	}
	return componentName, nil
}

func getPatchTypeFromFilename(fileName string) (types.PatchType, error) {
	idxDot := strings.LastIndex(fileName, ".")
	if idxDot == -1 {
		// Shift the dot index into the length of the string.
		idxDot = len(fileName)
	}

	idxPlus := strings.Index(fileName, "+")
	if idxPlus == -1 || idxPlus > idxDot {
		// If the + character is missing or if its index is after the dot,
		// just shift it into the dot index.
		idxPlus = idxDot
	} else {
		// Increment the plus index to discard the + character itself.
		idxPlus++
	}

	patchType := fileName[idxPlus:idxDot]
	pt, ok := patchTypes[patchType]
	if !ok {
		return "", errors.Errorf("unknown patch type %q from file name %q", patchType, fileName)
	}
	return pt, nil
}

// createPatchSet creates a patchSet object.
func createPatchSet(componentName string, patchType types.PatchType, data string) (*patchSet, error) {
	var patches []string

	if len(data) > 0 {
		// Split the patches and convert them to JSON.
		patches = strings.Split(data, "\n---")
		for i, patch := range patches {
			patchJSON, err := yaml.YAMLToJSON([]byte(patch))
			if err != nil {
				return nil, errors.Wrapf(err, "could not convert patch to JSON:\n%s\n", patch)
			}
			patches[i] = string(patchJSON)
		}
	}

	// Create patchSet.
	return &patchSet{
		componentName: componentName,
		patchType:     patchType,
		patches:       patches,
	}, nil
}

func filerPatchFilesForPath(targetPath string) ([]*patchFile, []string, error) {
	var ignoredFiles []string
	var patchFiles []*patchFile

	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		// Directories are ignored.
		if info.IsDir() {
			return nil
		}

		// Only support the .yaml and .json extensions.
		if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".json" {
			ignoredFiles = append(ignoredFiles, path)
			return nil
		}

		// Get the component name from the filename; else print a warning and skip.
		componentName, err := getComponentNameFromFilename(info.Name())
		if err != nil {
			ignoredFiles = append(ignoredFiles, path)
			return nil
		}

		pf := &patchFile{
			path:          path,
			componentName: componentName,
		}
		patchFiles = append(patchFiles, pf)
		return nil
	})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not list files for path %q", targetPath)
	}
	return patchFiles, ignoredFiles, nil
}

func createPatchSetsFromPatchFiles(files []*patchFile) ([]*patchSet, error) {
	patchSets := make([]*patchSet, 0, len(files))
	for _, f := range files {
		fileName := filepath.Base(f.path)

		// Read the patch file.
		data, err := ioutil.ReadFile(f.path)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read patches from file %q", f)
		}

		// Get the patch type from the filename.
		patchType, err := getPatchTypeFromFilename(fileName)
		if err != nil {
			return nil, err
		}

		// Create a patchSet object.
		patchSet, err := createPatchSet(f.componentName, patchType, string(data))
		if err != nil {
			return nil, err
		}
		patchSets = append(patchSets, patchSet)
	}

	return patchSets, nil
}

func createStaticPodPatchTargets(path string) ([]*componentPatchTarget, error) {
	var targets []*componentPatchTarget

	for _, component := range components {
		componentFile := filepath.Join(path, component, ".yaml")

		data, err := ioutil.ReadFile(componentFile)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read component file %q", componentFile)
		}

		target := &componentPatchTarget{
			name:                 component,
			data:                 data,
			strategicPatchObject: v1.Pod{},
		}
		targets = append(targets, target)
	}

	return targets, nil
}

func writeStaticPodTargets(path string, targets []*componentPatchTarget) error {
	for _, target := range targets {
		componentFile := filepath.Join(path, target.name, ".yaml")
		if err := ioutil.WriteFile(componentFile, target.data, 0644); err != nil {
			return errors.Wrapf(err, "could not write component file %q", componentFile)
		}
	}
	return nil
}

func applyPatchSetsToComponentTargets(componentTargets []*componentPatchTarget, patchSets []*patchSet) error {
	// Iterate over the component targets.
	for i, componentTarget := range componentTargets {
		var err error
		var patchedData []byte

		// Always convert the component data to JSON.
		patchedData, err = yaml.YAMLToJSON(componentTarget.data)
		if err != nil {
			return err
		}

		// Iterate over the patchSets.
		for _, patchSet := range patchSets {
			if patchSet.componentName != componentTarget.name {
				continue
			}

			// Iterate over the patches in the patchSets.
			for _, patch := range patchSet.patches {
				patchBytes := []byte(patch)

				// Patch based on the patch type.
				switch patchSet.patchType {

				// JSON patch.
				case types.JSONPatchType:
					var patchObj jsonpatch.Patch
					patchObj, err = jsonpatch.DecodePatch(patchBytes)
					if err == nil {
						patchedData, err = patchObj.Apply(patchedData)
					}

				// Merge patch.
				case types.MergePatchType:
					patchedData, err = jsonpatch.MergePatch(patchedData, patchBytes)

				// Strategic merge patch.
				case types.StrategicMergePatchType:
					patchedData, err = strategicpatch.StrategicMergePatch(
						patchedData,
						patchBytes,
						componentTarget.strategicPatchObject,
					)
				}

				if err != nil {
					return errors.Wrapf(err, "could not apply the following patch of type %q to component %q:\n%s\n",
						patchSet.patchType,
						componentTarget.name,
						patch)
				}
				fmt.Printf("[patches] applied patch of type %q to component %q\n", patchSet.patchType, componentTarget.name)
			}

			// Convert the data back to YAML.
			patchedData, err = yaml.JSONToYAML(patchedData)
			if err != nil {
				return err
			}

			// Update the data for this component target.
			componentTargets[i].data = patchedData
		}
	}

	return nil
}

func process(componentPath string, patchesPath string) error {
	// Get the files in the path.
	patchFiles, ignoredFiles, err := filerPatchFilesForPath(patchesPath)
	if err != nil {
		return err
	}

	fmt.Printf("[patches] found the following patch files in path %q: %v\n", patchesPath, patchFiles)
	fmt.Printf("[patches] ignored the following files in path %q: %v\n", patchesPath, ignoredFiles)

	// Read the patches from the given path.
	patchSets, err := createPatchSetsFromPatchFiles(patchFiles)
	if err != nil {
		return err
	}

	// Reads the static Pods from disk.
	componentTargets, err := createStaticPodPatchTargets(componentPath)
	if err != nil {
		return err
	}

	// Apply the patches to the component targets.
	if err := applyPatchSetsToComponentTargets(componentTargets, patchSets); err != nil {
		return err
	}

	// Write the static Pods to disk.
	if err := writeStaticPodTargets(componentPath, componentTargets); err != nil {
		return err
	}
	return nil
}

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
