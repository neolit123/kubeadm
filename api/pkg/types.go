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

// Kind respresents an interface that all API objects that support validation, defaulting
// and conversion in a version package must implement.
type Kind interface {
	// ConvertUp must take and older versioned Kind as input and covert it to a
	// Kind of the current version. Older version packages must not import newer ones.
	// Newer version packages must only import the prior version.
	ConvertUp(*Converter, Kind) (Kind, error)
	// ConvertDown must take the current version Kind as input and down-convert it
	// to a prior version Kind.
	ConvertDown(*Converter, Kind) (Kind, error)
	// ConvertUpName must return the Kind.Name() of this object in the prior version.
	ConvertUpName() string
	// Validate must define the validation function for this Kind.
	Validate() error
	// Default must define the defaulting function for this Kind.
	Default() error
	// TypeMeta ...
	GetTypeMeta() *metav1.TypeMeta
	// GetDefaultTypeMeta ...
	GetDefaultTypeMeta() *metav1.TypeMeta
}

// VersionKinds can be used to map Kinds to a version.
type VersionKinds struct {
	Version string
	Kinds   []Kind
}

// Group ...
type Group struct {
	Name     string
	Versions []VersionKinds
}
