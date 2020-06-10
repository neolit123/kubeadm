package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"sigs.k8s.io/yaml"
)

// https://github.com/kubernetes/kubernetes/blob/053d46c3c0a3690b6f9afa1df08b8ef71115b026/staging/src/k8s.io/kubectl/pkg/cmd/patch/patch.go#L298-L330
// https://github.com/kubernetes-sigs/kind/blob/a944589ec78b53fe62b45e8890e45e8d6c078f53/pkg/cluster/internal/patch/resource.go
// https://kubernetes.io/docs/tasks/run-application/update-api-object-kubectl-patch/#use-a-strategic-merge-patch-to-update-a-deployment

/*
Path to a directory that contains files named "target[suffix][+patchtype][.extension]".
For example, "kube-apiserver0+merge.yaml" or just "kube-apiserver.json".
"patchtype" can be one of "strategic", "merge" or "json" and they match the patch formats
supported by kubectl. The default "patchtype" is "strategic". "extension" must be either
".json" or ".yaml". "suffix" is an optional string that can be used to determine
which patches are applied first.
*/

// PatchTarget defines a target to be patched, such as a control-plane static Pod.
type PatchTarget struct {
	// Name must be the name of a known target.
	Name string

	// StrategicMergePatchObject is only used for strategic merge patches.
	// It represents the underlying object type that is patched - e.g. "v1.Pod"
	StrategicMergePatchObject interface{}

	// Data must contain the bytes that will be patched.
	Data []byte
}

// PatchManager defines an object that can apply patches.
type PatchManager struct {
	patchSets    []*patchSet
	knownTargets []string
	output       io.Writer
}

// patchSet defines a set of patches of a certain type that can patch a PatchTarget.
type patchSet struct {
	targetName string
	patchType  types.PatchType
	patches    []string
}

// String() is used for unit-testing.
func (ps *patchSet) String() string {
	return fmt.Sprintf(
		"{%q, %q, %#v}",
		ps.targetName,
		ps.patchType,
		ps.patches,
	)
}

var (
	pathLock  = &sync.RWMutex{}
	pathCache = map[string]*PatchManager{}

	patchTypes = map[string]types.PatchType{
		"json":      types.JSONPatchType,
		"merge":     types.MergePatchType,
		"strategic": types.StrategicMergePatchType,
		"":          types.StrategicMergePatchType, // Treat an empty value as the default = strategic.
	}
)

// GetPatchManagerForPath creates a patch manager that can be used to apply patches to "knownTargets".
// "path" should contain patches that can be used to patch the "knownTargets".
// If "output" is non-nil, messages about actions performed by the manager would go on this io.Writer.
func GetPatchManagerForPath(path string, knownTargets []string, output io.Writer) (*PatchManager, error) {
	pathLock.RLock()
	if pm, known := pathCache[path]; known {
		pathLock.RUnlock()
		return pm, nil
	}
	pathLock.RUnlock()

	if output == nil {
		output = ioutil.Discard
	}

	fmt.Fprintf(output, "[patches] reading patches from path %q\n", path)

	// Get the files in the path.
	patchSets, ignoredFiles, err := getPatchSetsFromPath(path, knownTargets)
	if err != nil {
		return nil, err
	}

	if len(patchSets) > 0 {
		fmt.Fprintf(output, "[patches] found the following patch files: %v\n", patchSets)
	}
	if len(ignoredFiles) > 0 {
		fmt.Fprintf(output, "[patches] ignored the following files: %v\n", ignoredFiles)
	}

	pm := &PatchManager{
		patchSets:    patchSets,
		knownTargets: knownTargets,
		output:       output,
	}
	pathLock.Lock()
	pathCache[path] = pm
	pathLock.Unlock()

	return pm, nil
}

// ApplyPatchesToTarget takes a patch target and patches its "Data" using the patches
// stored in the patch manager. The resulted "Data" is always converted to JSON.
func (pm *PatchManager) ApplyPatchesToTarget(patchTarget *PatchTarget) error {
	var err error
	var patchedData []byte

	var found bool
	for _, pt := range pm.knownTargets {
		if pt == patchTarget.Name {
			found = true
			break
		}
	}
	if !found {
		return errors.Errorf("unknown patch target name %q, must be one of %v", patchTarget.Name, pm.knownTargets)
	}

	// Always convert the target data to JSON.
	patchedData, err = yaml.YAMLToJSON(patchTarget.Data)
	if err != nil {
		return err
	}

	// Iterate over the patchSets.
	for _, patchSet := range pm.patchSets {
		if patchSet.targetName != patchTarget.Name {
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
					patchTarget.StrategicMergePatchObject,
				)
			}

			if err != nil {
				return errors.Wrapf(err, "could not apply the following patch of type %q to target %q:\n%s\n",
					patchSet.patchType,
					patchTarget.Name,
					patch)
			}
			fmt.Fprintf(pm.output, "[patches] applied patch of type %q to target %q\n", patchSet.patchType, patchTarget.Name)
		}

		// Update the data for this patch target.
		patchTarget.Data = patchedData
	}

	return nil
}

// getTargetNameFromFilename accepts a file name and returns a known target name string,
// or an error if the target name is unknown.
func getTargetNameFromFilename(fileName string, knownTargets []string) (string, error) {
	var targetName string

	for _, target := range knownTargets {
		if strings.HasPrefix(fileName, target) {
			targetName = target
			break
		}
	}

	if len(targetName) == 0 {
		return "", errors.Errorf("received file name %q, but target must be one of %v", knownTargets, fileName)
	}
	return targetName, nil
}

// getPatchTypeFromFilename accepts a file name and returns the patch type encoded in it.
// For example, "etcd+merge.json" would return "merge". Returns an error on a unknown patch types.
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
func createPatchSet(targetName string, patchType types.PatchType, data string) (*patchSet, error) {
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
		targetName: targetName,
		patchType:  patchType,
		patches:    patches,
	}, nil
}

// getPatchSetsFromPath walks a path, ignores sub-directories and non-patch files, and
// returns a list of patchFile objects.
func getPatchSetsFromPath(targetPath string, knownTargets []string) ([]*patchSet, []string, error) {
	var ignoredFiles []string
	patchSets := []*patchSet{}

	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Directories are ignored.
		if info.IsDir() {
			return nil
		}

		fileName := filepath.Base(path)

		// Only support the .yaml and .json extensions.
		if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".json" {
			ignoredFiles = append(ignoredFiles, fileName)
			return nil
		}

		// Get the target name from the filename. If there is an error ignore the file.
		targetName, err := getTargetNameFromFilename(info.Name(), knownTargets)
		if err != nil {
			ignoredFiles = append(ignoredFiles, fileName)
			return nil
		}

		// Get the patch type from the filename.
		patchType, err := getPatchTypeFromFilename(fileName)
		if err != nil {
			return err
		}

		// Read the patch file.
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "could not read the file %q", path)
		}

		// Create a patchSet object.
		patchSet, err := createPatchSet(targetName, patchType, string(data))
		if err != nil {
			return err
		}

		patchSets = append(patchSets, patchSet)
		return nil
	})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not list files for path %q", targetPath)
	}

	return patchSets, ignoredFiles, nil
}

// PatchStaticPod patch a static Pod with patches stored in patchsDir.
func PatchStaticPod(pod *v1.Pod, patchsDir string, output io.Writer) (*v1.Pod, error) {
	// Marshal the Pod manifest into YAML.
	podYAML, err := yaml.Marshal(pod)
	if err != nil {
		return pod, errors.Wrapf(err, "failed to marshal Pod manifest to YAML")
	}

	var testKnownTargets = []string{
		"etcd",
		"kube-apiserver",
		"kube-controller-manager",
		"kube-scheduler",
	}

	patchManager, err := GetPatchManagerForPath(patchsDir, testKnownTargets, output)
	if err != nil {
		return pod, err
	}

	patchTarget := &PatchTarget{
		Name:                      pod.Name,
		StrategicMergePatchObject: v1.Pod{},
		Data:                      podYAML,
	}
	if err := patchManager.ApplyPatchesToTarget(patchTarget); err != nil {
		return pod, err
	}

	pod2 := &v1.Pod{}
	if err := yaml.Unmarshal(patchTarget.Data, pod2); err != nil {
		return pod, errors.Wrapf(err, "failed to unmarshal YAML manifest to Pod")
	}

	return pod2, nil
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

// KustomizeStaticPod applies patches defined in kustomizeDir to a static Pod manifest
// func KustomizeStaticPod(pod *v1.Pod, kustomizeDir string) (*v1.Pod, error) {
// 	// marshal the pod manifest into yaml
// 	serialized, err := kubeadmutil.MarshalToYaml(pod, v1.SchemeGroupVersion)
// 	if err != nil {
// 		return pod, errors.Wrapf(err, "failed to marshal manifest to YAML")
// 	}

// 	km, err := kustomize.GetManager(kustomizeDir)
// 	if err != nil {
// 		return pod, errors.Wrapf(err, "failed to GetPatches from %q", kustomizeDir)
// 	}

// 	kustomized, err := km.Kustomize(serialized)
// 	if err != nil {
// 		return pod, errors.Wrap(err, "failed to kustomize static Pod manifest")
// 	}

// 	// unmarshal kustomized yaml back into a pod manifest
// 	obj, err := kubeadmutil.UnmarshalFromYaml(kustomized, v1.SchemeGroupVersion)
// 	if err != nil {
// 		return pod, errors.Wrap(err, "failed to unmarshal kustomize manifest from YAML")
// 	}

// 	pod2, ok := obj.(*v1.Pod)
// 	if !ok {
// 		return pod, errors.Wrap(err, "kustomized manifest is not a valid Pod object")
// 	}

// 	return pod2, nil
// }
