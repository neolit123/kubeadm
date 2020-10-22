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
	"k8s.io/kubeadm/api/pkg"
	"k8s.io/kubeadm/api/v1beta1"
)

// ConvertUp ...
func (*InitConfiguration) ConvertUp(cv *pkg.Converter, in pkg.Kind) (pkg.Kind, error) {
	new := &InitConfiguration{}
	cv.DeepCopy(new, in)
	// restore from cache
	cachedKind := cv.GetFromCache(new)
	if cachedKind != nil {
		cached := cachedKind.(*InitConfiguration)
		new.CertificateKey = cached.CertificateKey
		new.NodeRegistration.IgnorePreflightErrors = make([]string, len(cached.NodeRegistration.IgnorePreflightErrors))
		copy(new.NodeRegistration.IgnorePreflightErrors, cached.NodeRegistration.IgnorePreflightErrors)
	}
	return new, nil
}

// ConvertDown ...
func (*InitConfiguration) ConvertDown(cv *pkg.Converter, in pkg.Kind) (pkg.Kind, error) {
	cv.AddToCache(in)
	new := &v1beta1.InitConfiguration{}
	cv.DeepCopy(new, in)
	return new, nil
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
func (*ClusterConfiguration) ConvertUp(cv *pkg.Converter, in pkg.Kind) (pkg.Kind, error) {
	new := &ClusterConfiguration{}
	cv.DeepCopy(new, in)
	return new, nil
}

// ConvertDown ...
func (*ClusterConfiguration) ConvertDown(cv *pkg.Converter, in pkg.Kind) (pkg.Kind, error) {
	new := &v1beta1.ClusterConfiguration{}
	cv.DeepCopy(new, in)
	return new, nil
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
func (*ClusterStatus) ConvertUp(cv *pkg.Converter, in pkg.Kind) (pkg.Kind, error) {
	return in, nil
}

// ConvertDown ...
func (*ClusterStatus) ConvertDown(cv *pkg.Converter, in pkg.Kind) (pkg.Kind, error) {
	return in, nil
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
func (*JoinConfiguration) ConvertUp(cv *pkg.Converter, in pkg.Kind) (pkg.Kind, error) {
	new := &JoinConfiguration{}
	cv.DeepCopy(new, in)
	// restore from cache
	cachedKind := cv.GetFromCache(new)
	if cachedKind != nil {
		cached := cachedKind.(*JoinConfiguration)
		new.NodeRegistration.IgnorePreflightErrors = make([]string, len(cached.NodeRegistration.IgnorePreflightErrors))
		copy(new.NodeRegistration.IgnorePreflightErrors, cached.NodeRegistration.IgnorePreflightErrors)
	}
	return new, nil
}

// ConvertDown ...
func (*JoinConfiguration) ConvertDown(cv *pkg.Converter, in pkg.Kind) (pkg.Kind, error) {
	cv.AddToCache(in)
	new := &v1beta1.JoinConfiguration{}
	cv.DeepCopy(new, in)
	return new, nil
}

// ConvertUpName ...
func (*JoinConfiguration) ConvertUpName() string {
	return "JoinConfiguration"
}

// ConvertDownName ...
func (*JoinConfiguration) ConvertDownName() string {
	return "JoinConfiguration"
}
