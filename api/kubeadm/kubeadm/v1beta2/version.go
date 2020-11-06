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

package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/kubeadm/api/kubeadm/kubeadm"
)

// Version ...
const Version = "v1beta2"

// GetDefaultTypeMeta ...
func (*InitConfiguration) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: kubeadm.Group + "/" + Version, Kind: "InitConfiguration"}
}

// ---------------

// GetDefaultTypeMeta ...
func (*ClusterConfiguration) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: kubeadm.Group + "/" + Version, Kind: "ClusterConfiguration"}
}

// ---------------

// GetDefaultTypeMeta ...
func (*ClusterStatus) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: kubeadm.Group + "/" + Version, Kind: "ClusterStatus"}
}

// ---------------

// GetDefaultTypeMeta ...
func (*JoinConfiguration) GetDefaultTypeMeta() *metav1.TypeMeta {
	return &metav1.TypeMeta{APIVersion: kubeadm.Group + "/" + Version, Kind: "JoinConfiguration"}
}
