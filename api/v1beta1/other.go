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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Version ...
const Version = "v1beta1"

// GetTypeMeta ...
func (x *InitConfiguration) GetTypeMeta() *metav1.TypeMeta {
	return &x.TypeMeta
}

// Name ...
func (*InitConfiguration) Name() string {
	return "InitConfiguration"
}

// Version ...
func (*InitConfiguration) Version() string {
	return Version
}

// ---------------

// GetTypeMeta ...
func (x *ClusterConfiguration) GetTypeMeta() *metav1.TypeMeta {
	return &x.TypeMeta
}

// Name ...
func (*ClusterConfiguration) Name() string {
	return "ClusterConfiguration"
}

// Version ...
func (*ClusterConfiguration) Version() string {
	return Version
}

// ---------------

// GetTypeMeta ...
func (x *ClusterStatus) GetTypeMeta() *metav1.TypeMeta {
	return &x.TypeMeta
}

// Name ...
func (*ClusterStatus) Name() string {
	return "ClusterStatus"
}

// Version ...
func (*ClusterStatus) Version() string {
	return Version
}

// ---------------

// GetTypeMeta ...
func (x *JoinConfiguration) GetTypeMeta() *metav1.TypeMeta {
	return &x.TypeMeta
}

// Name ...
func (*JoinConfiguration) Name() string {
	return "JoinConfiguration"
}

// Version ...
func (*JoinConfiguration) Version() string {
	return Version
}
