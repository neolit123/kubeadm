package shared

// import (
// 	"encoding/json"
// 	"fmt"
// 	"strings"

// 	// "k8s.io/kubeadm/api/scheme"
// )

// // IsKnownGroup ...
// func IsKnownGroup(g string) bool {
// 	return g == scheme.Group
// }

// // Group ...
// func Group() string {
// 	return scheme.Group
// }

// // VersionsKinds ...
// func VersionsKinds() map[string][]string {
// 	return versionKinds
// }

// // IsKnownGroupVersion ...
// func IsKnownGroupVersion(group, version string) bool {
// 	return isKnownGroupVersion(scheme.Group, versionKinds, group, version)
// }

// func isKnownGroupVersion(knownGroup string, knownVersionKinds map[string][]string, group, version string) bool {
// 	if group != knownGroup {
// 		return false
// 	}
// 	if _, known := knownVersionKinds[version]; !known {
// 		return false
// 	}
// 	return true
// }

// // IsKnownVersionKind ...
// func IsKnownVersionKind(version, kind string) bool {
// 	return isKnownVersionKind(versionKinds, version, kind)
// }

// func isKnownVersionKind(knownVersionKinds map[string][]string, version, kind string) bool {
// 	kinds, known := knownVersionKinds[version]
// 	if !known {
// 		return false
// 	}
// 	for _, k := range kinds {
// 		if k == kind {
// 			return true
// 		}
// 	}
// 	return false
// }

// // ParseAPIVersion ...
// func ParseAPIVersion(apiVersion string) (group, version string, err error) {
// 	err = fmt.Errorf("badly formatted apiVersion %q, must be 'group/version'", apiVersion)
// 	gv := strings.Split(apiVersion, "/")
// 	if len(gv) != 2 {
// 		return "", "", err
// 	}
// 	if len(gv[0]) == 0 || len(gv[1]) == 0 {
// 		return "", "", err
// 	}
// 	return gv[0], gv[1], nil
// }

// // ReadAPIVersionKindFromJSON ...
// func ReadAPIVersionKindFromJSON(in []byte) (apiVersion, kind string, err error) {
// 	const errorFormat = "cannot extract TypeMeta from input"
// 	tm := &TypeMeta{}
// 	if err = json.Unmarshal(in, tm); err != nil {
// 		return "", "", fmt.Errorf("%s: %w", errorFormat, err)
// 	}
// 	if len(tm.APIVersion) == 0 {
// 		return "", "", fmt.Errorf("%s: empty apiVersion", errorFormat)
// 	}
// 	if len(tm.Kind) == 0 {
// 		return "", "", fmt.Errorf("%s: empty kind", errorFormat)
// 	}
// 	return tm.APIVersion, tm.Kind, nil
// }
