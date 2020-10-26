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

package shared

const (
	// IPv6DualStack is expected to be alpha in v1.16
	IPv6DualStack = "IPv6DualStack"
	// PublicKeysECDSA is expected to be alpha in v1.19
	PublicKeysECDSA = "PublicKeysECDSA"
)

// FeatureEnabled indicates whether a feature name has been enabled
func FeatureEnabled(featureList map[string]bool, featureName string) bool {
	if enabled, ok := featureList[string(featureName)]; ok {
		return enabled
	}
	return false
}
