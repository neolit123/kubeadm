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

// ConvertFrom ...
func (*InitConfiguration) ConvertFrom(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	return in, nil
}

// ConvertTo ...
func (*InitConfiguration) ConvertTo(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	return in, nil
}

// ConvertFromSpec ...
func (*InitConfiguration) ConvertFromSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&InitConfiguration{})
}

// ConvertToSpec ...
func (*InitConfiguration) ConvertToSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&InitConfiguration{})
}

// -------

// ConvertFrom ...
func (*ClusterConfiguration) ConvertFrom(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	return in, nil
}

// ConvertTo ...
func (*ClusterConfiguration) ConvertTo(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	return in, nil
}

// ConvertFromSpec ...
func (*ClusterConfiguration) ConvertFromSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&ClusterConfiguration{})
}

// ConvertToSpec ...
func (*ClusterConfiguration) ConvertToSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&ClusterConfiguration{})
}

// -------

// ConvertFrom ...
func (*ClusterStatus) ConvertFrom(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	return in, nil
}

// ConvertTo ...
func (*ClusterStatus) ConvertTo(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	return in, nil
}

// ConvertFromSpec ...
func (*ClusterStatus) ConvertFromSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&ClusterStatus{})
}

// ConvertToSpec ...
func (*ClusterStatus) ConvertToSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&ClusterStatus{})
}

// -------

// ConvertFrom ...
func (*JoinConfiguration) ConvertFrom(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	return in, nil
}

// ConvertTo ...
func (*JoinConfiguration) ConvertTo(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	return in, nil
}

// ConvertFromSpec ...
func (*JoinConfiguration) ConvertFromSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&JoinConfiguration{})
}

// ConvertToSpec ...
func (*JoinConfiguration) ConvertToSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&JoinConfiguration{})
}
