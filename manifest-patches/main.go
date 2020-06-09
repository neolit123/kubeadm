package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/yaml"
)

// https://github.com/kubernetes/kubernetes/blob/053d46c3c0a3690b6f9afa1df08b8ef71115b026/staging/src/k8s.io/kubectl/pkg/cmd/patch/patch.go#L298-L330
// https://github.com/kubernetes-sigs/kind/blob/a944589ec78b53fe62b45e8890e45e8d6c078f53/pkg/cluster/internal/patch/resource.go
// https://kubernetes.io/docs/tasks/run-application/update-api-object-kubectl-patch/#use-a-strategic-merge-patch-to-update-a-deployment

/*
Path to a folder that contains files named "componentname[index][+patchtype][.extension]".
For example "kube-apiserver0+merge.yaml" or just "kube-apiserver.yaml". "index" is a number starting
from zero and determines which patch is applied first. "patchtype" can be one of "strategic", "merge"
or "json" and they match the patch formats supported by kubectl. The default "patchtype" is "strategic".
*/

var patchTypes = map[string]types.PatchType{
	"json":      types.JSONPatchType,
	"merge":     types.MergePatchType,
	"strategic": types.StrategicMergePatchType,
	"":          types.StrategicMergePatchType, // Treat an empty value as the default = strategic.
}

// ComponentTarget defines a component target to be patched.
type ComponentTarget struct {
	// ComponentName must be a known component name or identifier - e.g. "etcd"
	ComponentName string

	// StrategicMergePatchObject is only used for strategic merge patches.
	// It represents the underlying Kubernetes object type that is patched - e.g. "v1.Pod"
	StrategicMergePatchObject interface{}

	// Data must contain the bytes that will be patched.
	Data []byte
}

// PatchManager defines a patch manager that holds a set of patches that can be applied to a target.
type PatchManager struct {
	patchSets []*patchSet
	output    io.Writer
}

// patchFile defines a set of patches of a certain type that target a component.
type patchSet struct {
	componentName string
	patchType     types.PatchType
	patches       []string
}

// patchFile is a utility structure used when reading patch files from a path.
type patchFile struct {
	componentName string
	path          string
	data          string
}

// String() is used for unit-testing.
func (ps *patchSet) String() string {
	return fmt.Sprintf(
		"{%q, %q, %q}",
		ps.componentName,
		ps.patchType,
		ps.patches,
	)
}

// String() is used for unit-testing.
func (pf *patchFile) String() string {
	return fmt.Sprintf(
		"{%q, %q, %q}",
		pf.componentName,
		pf.path,
		pf.data,
	)
}

// GetPatchManagerFromPath creates a patch manager that can be used to apply patches to "knownTargets".
// "path" should contain patches that can be used to patch the "knownTargets".
// If "output" is non-nil, messages about actions performed by the manager would go on this io.Writer.
func GetPatchManagerFromPath(path string, knownTargets []string, output io.Writer) (*PatchManager, error) {
	if output == nil {
		output = ioutil.Discard
	}

	fmt.Fprintf(output, "[patches] reading patches from path %q\n", path)

	// Get the files in the path.
	patchFiles, ignoredFiles, err := getPatchFilesForPath(path, knownTargets)
	if err != nil {
		return nil, err
	}

	if len(patchFiles) > 0 {
		fmt.Fprintf(output, "[patches] found the following patch files: %v\n", patchFiles)
	}
	if len(ignoredFiles) > 0 {
		fmt.Fprintf(output, "[patches] ignored the following files: %v\n", patchFiles)
	}

	// Read the patches from the given path.
	patchSets, err := createPatchSetsFromPatchFiles(patchFiles)
	if err != nil {
		return nil, err
	}

	return &PatchManager{patchSets: patchSets, output: output}, nil
}

// PatchComponentTarget takes a component target and patches its "Data" it using the patch
// sets stored in the patch manager. The resulted "Data" is always converted to JSON.
func (pm *PatchManager) PatchComponentTarget(componentTarget *ComponentTarget) error {
	var err error
	var patchedData []byte

	// Always convert the component data to JSON.
	patchedData, err = yaml.YAMLToJSON(componentTarget.Data)
	if err != nil {
		return err
	}

	// Iterate over the patchSets.
	for _, patchSet := range pm.patchSets {
		if patchSet.componentName != componentTarget.ComponentName {
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
					componentTarget.StrategicMergePatchObject,
				)
			}

			if err != nil {
				return errors.Wrapf(err, "could not apply the following patch of type %q to component %q:\n%s\n",
					patchSet.patchType,
					componentTarget.ComponentName,
					patch)
			}
			fmt.Fprintf(pm.output, "[patches] applied patch of type %q to component %q\n", patchSet.patchType, componentTarget.ComponentName)
		}

		// Update the data for this component target.
		componentTarget.Data = patchedData
	}

	return nil
}

// getComponentNameFromFilename accepts a file name and returns a known component string,
// or an error if the component is unknown.
func getComponentNameFromFilename(fileName string, knownTargets []string) (string, error) {
	var componentName string

	for _, c := range knownTargets {
		if strings.HasPrefix(fileName, c) {
			componentName = c
			break
		}
	}

	if len(componentName) == 0 {
		return "", errors.Errorf("component name must be one of %v from file name %q", knownTargets, fileName)
	}
	return componentName, nil
}

// getPatchTypeFromFilename accepts a file name and returns the patch type encoded in it.
// For example, "etcd+merge.json" would return "merge". Returns an  error on a unknown patch types.
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

// createPatchSet creates a patchSet object, by splitting the given "data" by "\n---".
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

	return &patchSet{
		componentName: componentName,
		patchType:     patchType,
		patches:       patches,
	}, nil
}

// getPatchFilesForPath walks a path, ignores sub-directories and non-patch files, and
// returns a list of patchFile objects.
func getPatchFilesForPath(targetPath string, knownTargets []string) ([]*patchFile, []string, error) {
	var ignoredFiles []string
	var patchFiles []*patchFile

	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Directories are ignored.
		if info.IsDir() {
			return nil
		}

		// Only support the .yaml and .json extensions.
		if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".json" {
			ignoredFiles = append(ignoredFiles, path)
			return nil
		}

		// Get the component name from the filename. If there is an error ignore the file.
		componentName, err := getComponentNameFromFilename(info.Name(), knownTargets)
		if err != nil {
			ignoredFiles = append(ignoredFiles, path)
			return nil
		}

		// Read the patch file.
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "could not read the file %q", path)
		}

		pf := &patchFile{
			path:          path,
			data:          string(data),
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

// createPatchSetsFromPatchFiles takes a list of patchFile objects and returns a list of patchSet
// objects. The function also ensures that unknown patch types encoded in a filename throw errors
// and are not ignored.
func createPatchSetsFromPatchFiles(files []*patchFile) ([]*patchSet, error) {
	patchSets := make([]*patchSet, 0, len(files))

	for _, f := range files {
		fileName := filepath.Base(f.path)

		// Get the patch type from the filename.
		patchType, err := getPatchTypeFromFilename(fileName)
		if err != nil {
			return nil, err
		}

		// Create a patchSet object.
		patchSet, err := createPatchSet(f.componentName, patchType, string(f.data))
		if err != nil {
			return nil, err
		}
		patchSets = append(patchSets, patchSet)
	}

	return patchSets, nil
}

func process(componentPath string, path string) error {
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
