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
	"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ConverterCacheAnnotation ...
	ConverterCacheAnnotation = "clusterlifecycle.x-k8s.io/converter-cache"
)

// Kind respresents an interface that all API objects that support validation, defaulting
// and conversion in a version package must implement.
type Kind interface {
	// ConvertUp must take and older versioned Kind as input and covert it to a
	// Kind of the current version. Older version packages must not import newer ones.
	// Newer version packages must only import the prior version.
	ConvertUp(*Converter, *KindSpec) (*KindSpec, error)
	// ConvertDown must take the current version Kind as input and down-convert it
	// to a prior version Kind.
	ConvertDown(*Converter, *KindSpec) (*KindSpec, error)
	// ConvertUpSpec ...
	ConvertUpSpec() *KindSpec
	// ConvertDownSpec ...
	ConvertDownSpec() *KindSpec
	// Validate must define the validation function for this Kind.
	Validate() error
	// Default must define the defaulting function for this Kind.
	Default() error
	// GetDefaultTypeMeta ...
	GetDefaultTypeMeta() *metav1.TypeMeta
}

// Version can be used to map Kinds to a version.
type Version struct {
	Version    string
	AddedIn    string
	Preferred  bool
	Deprecated bool
	Kinds      []Kind
}

// Group ...
type Group struct {
	Group      string
	AddedIn    string
	Deprecated bool
	Versions   []Version
}

// Converter ...
type Converter struct {
	groups        []Group
	cache         map[string]Kind
	output        io.Writer
	unmarshalFunc func([]byte, interface{}) error
	marshalFunc   func(interface{}) ([]byte, error)
}

// KindSpec ...
type KindSpec struct {
	Kinds []Kind
}
