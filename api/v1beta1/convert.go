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
	"k8s.io/kubeadm/api/pkg"
)

// ConvertUp ...
func (*InitConfiguration) ConvertUp(*pkg.Converter, pkg.Kind) (pkg.Kind, error) {
	return nil, nil
}

// ConvertDown ...
func (*InitConfiguration) ConvertDown(*pkg.Converter, pkg.Kind) (pkg.Kind, error) {
	return nil, nil
}

// ConvertUpName ...
func (*InitConfiguration) ConvertUpName() string {
	return "InitConfiguration"
}

// ConvertDownName ...
func (*InitConfiguration) ConvertDownName() string {
	return "InitConfiguration"
}

// -------

// ConvertUp ...
func (*ClusterConfiguration) ConvertUp(*pkg.Converter, pkg.Kind) (pkg.Kind, error) {
	return nil, nil
}

// ConvertDown ...
func (*ClusterConfiguration) ConvertDown(*pkg.Converter, pkg.Kind) (pkg.Kind, error) {
	return nil, nil
}

// ConvertUpName ...
func (*ClusterConfiguration) ConvertUpName() string {
	return "ClusterConfiguration"
}

// ConvertDownName ...
func (*ClusterConfiguration) ConvertDownName() string {
	return "ClusterConfiguration"
}

// -------

// ConvertUp ...
func (*ClusterStatus) ConvertUp(*pkg.Converter, pkg.Kind) (pkg.Kind, error) {
	return nil, nil
}

// ConvertDown ...
func (*ClusterStatus) ConvertDown(*pkg.Converter, pkg.Kind) (pkg.Kind, error) {
	return nil, nil
}

// ConvertUpName ...
func (*ClusterStatus) ConvertUpName() string {
	return "ClusterStatus"
}

// ConvertDownName ...
func (*ClusterStatus) ConvertDownName() string {
	return "ClusterStatus"
}

// -------

// ConvertUp ...
func (*JoinConfiguration) ConvertUp(*pkg.Converter, pkg.Kind) (pkg.Kind, error) {
	return nil, nil
}

// ConvertDown ...
func (*JoinConfiguration) ConvertDown(*pkg.Converter, pkg.Kind) (pkg.Kind, error) {
	return nil, nil
}

// ConvertUpName ...
func (*JoinConfiguration) ConvertUpName() string {
	return "JoinConfiguration"
}

// ConvertDownName ...
func (*JoinConfiguration) ConvertDownName() string {
	return "JoinConfiguration"
}
