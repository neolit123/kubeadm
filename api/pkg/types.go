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

package pkg

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Kind ...
type Kind interface {
	Version() string
	Name() string
	ConvertUp(*Converter, Kind) (Kind, error)
	ConvertDown(*Converter, Kind) (Kind, error)
	ConvertUpName() string
	ConvertDownName() string
	Validate() error
	Default() error
	GetTypeMeta() *metav1.TypeMeta
}

// VersionKinds ...
type VersionKinds struct {
	Version string
	Kinds   []Kind
}
