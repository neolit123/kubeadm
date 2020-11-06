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

package kubeadm

import (
	group "k8s.io/kubeadm/api/kubeadm/kubeadm"
	"k8s.io/kubeadm/api/kubeadm/kubeadm/v1beta1"
	"k8s.io/kubeadm/api/kubeadm/kubeadm/v1beta2"
	"k8s.io/kubeadm/api/pkg"
)

// Groups ...
var Groups = []pkg.Group{
	pkg.Group{
		Group:      group.Group,
		AddedIn:    "v1.5.0",
		Deprecated: false,
		Versions: []pkg.Version{
			{
				Version:    v1beta1.Version,
				AddedIn:    "v1.13.0",
				Preferred:  false,
				Deprecated: true,
				Kinds: []pkg.Kind{
					&v1beta1.InitConfiguration{},
					&v1beta1.ClusterConfiguration{},
					&v1beta1.ClusterStatus{},
					&v1beta1.JoinConfiguration{},
				},
			},
			{
				Version:    v1beta2.Version,
				AddedIn:    "v1.15.0",
				Preferred:  true,
				Deprecated: false,
				Kinds: []pkg.Kind{
					&v1beta2.InitConfiguration{},
					&v1beta2.ClusterConfiguration{},
					&v1beta2.ClusterStatus{},
					&v1beta2.JoinConfiguration{},
				},
			},
		},
	},
}
