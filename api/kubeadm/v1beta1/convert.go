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
func (*InitConfiguration) ConvertUp(cv *pkg.Converter, in *pkg.ConvertSpec) (*pkg.ConvertSpec, error) {
	return in, nil
}

// ConvertDown ...
func (*InitConfiguration) ConvertDown(cv *pkg.Converter, in *pkg.ConvertSpec) (*pkg.ConvertSpec, error) {
	return in, nil
}

// ConvertUpSpec ...
func (*InitConfiguration) ConvertUpSpec() *pkg.ConvertSpec {
	return &pkg.ConvertSpec{Kinds: []pkg.Kind{&InitConfiguration{}}}
}

// ConvertDownSpec ...
func (*InitConfiguration) ConvertDownSpec() *pkg.ConvertSpec {
	return &pkg.ConvertSpec{Kinds: []pkg.Kind{&InitConfiguration{}}}
}

// -------

// ConvertUp ...
func (*ClusterConfiguration) ConvertUp(cv *pkg.Converter, in *pkg.ConvertSpec) (*pkg.ConvertSpec, error) {
	return in, nil
}

// ConvertDown ...
func (*ClusterConfiguration) ConvertDown(cv *pkg.Converter, in *pkg.ConvertSpec) (*pkg.ConvertSpec, error) {
	return in, nil
}

// ConvertUpSpec ...
func (*ClusterConfiguration) ConvertUpSpec() *pkg.ConvertSpec {
	return &pkg.ConvertSpec{Kinds: []pkg.Kind{&ClusterConfiguration{}}}
}

// ConvertDownSpec ...
func (*ClusterConfiguration) ConvertDownSpec() *pkg.ConvertSpec {
	return &pkg.ConvertSpec{Kinds: []pkg.Kind{&ClusterConfiguration{}}}
}

// -------

// ConvertUp ...
func (*ClusterStatus) ConvertUp(cv *pkg.Converter, in *pkg.ConvertSpec) (*pkg.ConvertSpec, error) {
	return in, nil
}

// ConvertDown ...
func (*ClusterStatus) ConvertDown(cv *pkg.Converter, in *pkg.ConvertSpec) (*pkg.ConvertSpec, error) {
	return in, nil
}

// ConvertUpSpec ...
func (*ClusterStatus) ConvertUpSpec() *pkg.ConvertSpec {
	return &pkg.ConvertSpec{Kinds: []pkg.Kind{&ClusterStatus{}}}
}

// ConvertDownSpec ...
func (*ClusterStatus) ConvertDownSpec() *pkg.ConvertSpec {
	return &pkg.ConvertSpec{Kinds: []pkg.Kind{&ClusterStatus{}}}
}

// -------

// ConvertUp ...
func (*JoinConfiguration) ConvertUp(cv *pkg.Converter, in *pkg.ConvertSpec) (*pkg.ConvertSpec, error) {
	return in, nil
}

// ConvertDown ...
func (*JoinConfiguration) ConvertDown(cv *pkg.Converter, in *pkg.ConvertSpec) (*pkg.ConvertSpec, error) {
	return in, nil
}

// ConvertUpSpec ...
func (*JoinConfiguration) ConvertUpSpec() *pkg.ConvertSpec {
	return &pkg.ConvertSpec{Kinds: []pkg.Kind{&JoinConfiguration{}}}
}

// ConvertDownSpec ...
func (*JoinConfiguration) ConvertDownSpec() *pkg.ConvertSpec {
	return &pkg.ConvertSpec{Kinds: []pkg.Kind{&JoinConfiguration{}}}
}
