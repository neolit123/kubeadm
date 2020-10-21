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
