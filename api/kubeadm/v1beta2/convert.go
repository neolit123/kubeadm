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
	"k8s.io/kubeadm/api/kubeadm/v1beta1"
	"k8s.io/kubeadm/api/pkg"
)

// ConvertFrom ...
func (*InitConfiguration) ConvertFrom(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	ink := in.Kinds[0]
	new := &InitConfiguration{}
	pkg.DeepCopy(new, ink)
	// restore from cache
	cachedKind := cv.KindFromCache(new)
	if cachedKind != nil {
		cached := cachedKind.(*InitConfiguration)
		new.CertificateKey = cached.CertificateKey
		new.NodeRegistration.IgnorePreflightErrors = make([]string, len(cached.NodeRegistration.IgnorePreflightErrors))
		copy(new.NodeRegistration.IgnorePreflightErrors, cached.NodeRegistration.IgnorePreflightErrors)
	}
	return pkg.NewKindSpec().WithKinds(new), nil
}

// ConvertTo ...
func (*InitConfiguration) ConvertTo(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	ink := in.Kinds[0]
	cv.AddKindsToCache(ink)
	new := &v1beta1.InitConfiguration{}
	pkg.DeepCopy(new, ink)
	return pkg.NewKindSpec().WithKinds(new), nil
}

// ConvertFromSpec ...
func (*InitConfiguration) ConvertFromSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&v1beta1.InitConfiguration{})
}

// ConvertToSpec ...
func (*InitConfiguration) ConvertToSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&InitConfiguration{})
}

// -------

// ConvertFrom ...
func (*ClusterConfiguration) ConvertFrom(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	ink := in.Kinds[0]
	new := &ClusterConfiguration{}
	pkg.DeepCopy(new, ink)
	return pkg.NewKindSpec().WithKinds(new), nil
}

// ConvertTo ...
func (*ClusterConfiguration) ConvertTo(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	ink := in.Kinds[0]
	new := &v1beta1.ClusterConfiguration{}
	pkg.DeepCopy(new, ink)
	return pkg.NewKindSpec().WithKinds(new), nil
}

// ConvertFromSpec ...
func (*ClusterConfiguration) ConvertFromSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&v1beta1.ClusterConfiguration{})
}

// ConvertToSpec ...
func (*ClusterConfiguration) ConvertToSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&ClusterConfiguration{})
}

// -------

// ConvertFrom ...
func (*ClusterStatus) ConvertFrom(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	new := &ClusterStatus{}
	pkg.DeepCopy(new, in.Kinds[0])
	return pkg.NewKindSpec().WithKinds(new), nil
}

// ConvertTo ...
func (*ClusterStatus) ConvertTo(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	new := &v1beta1.ClusterStatus{}
	pkg.DeepCopy(new, in.Kinds[0])
	return pkg.NewKindSpec().WithKinds(new), nil
}

// ConvertFromSpec ...
func (*ClusterStatus) ConvertFromSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&v1beta1.ClusterStatus{})
}

// ConvertToSpec ...
func (*ClusterStatus) ConvertToSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&ClusterStatus{})
}

// -------

// ConvertFrom ...
func (*JoinConfiguration) ConvertFrom(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	ink := in.Kinds[0]
	new := &JoinConfiguration{}
	pkg.DeepCopy(new, ink)
	// restore from cache
	cachedKind := cv.KindFromCache(new)
	if cachedKind != nil {
		cached := cachedKind.(*JoinConfiguration)
		new.NodeRegistration.IgnorePreflightErrors = make([]string, len(cached.NodeRegistration.IgnorePreflightErrors))
		copy(new.NodeRegistration.IgnorePreflightErrors, cached.NodeRegistration.IgnorePreflightErrors)
	}
	return pkg.NewKindSpec().WithKinds(new), nil
}

// ConvertTo ...
func (*JoinConfiguration) ConvertTo(cv *pkg.Converter, in *pkg.KindSpec) (*pkg.KindSpec, error) {
	ink := in.Kinds[0]
	cv.AddKindsToCache(ink)
	new := &v1beta1.JoinConfiguration{}
	pkg.DeepCopy(new, ink)
	return pkg.NewKindSpec().WithKinds(new), nil
}

// ConvertFromSpec ...
func (*JoinConfiguration) ConvertFromSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&v1beta1.JoinConfiguration{})
}

// ConvertToSpec ...
func (*JoinConfiguration) ConvertToSpec() *pkg.KindSpec {
	return pkg.NewKindSpec().WithKinds(&JoinConfiguration{})
}
