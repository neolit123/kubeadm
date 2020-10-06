package v1beta1

import (
	"fmt"
	"k8s.io/kubeadm/api/shared"
)

const Version = "v1beta1"

// Foo ...
type Foo struct {
	shared.TypeMeta `json:",inline"`
	// A ...
	A string `json:"a,omitempty"`
	// B ...
	B string `json:"b,omitempty"`
	// C ...
	C string `json:"c,omitempty"`
}

var _ shared.Kind = (*Foo)(nil)

// Version ...
func (*Foo) Version() string {
	return Version
}

// Name ...
func (*Foo) Name() string {
	return "Foo"
}

// ConvertUp ...
func (*Foo) ConvertUp(cv *shared.Converter, in shared.Kind) (shared.Kind, error) {
	return in, nil
}

// ConvertDown ...
func (*Foo) ConvertDown(cv *shared.Converter, in shared.Kind) (shared.Kind, error) {
	return in, nil
}

// ConvertUpName ...
func (*Foo) ConvertUpName() string {
	return "Foo"
}

// ConvertDownName ...
func (*Foo) ConvertDownName() string {
	return "Foo"
}

// Validate ...
func (x *Foo) Validate() error {
	if len(x.C) == 0 {
		return fmt.Errorf("Foo.C cannot be empty")
	}
	return nil
}

// Default ...
func (x *Foo) Default() error {
	if len(x.C) == 0 {
		x.C = "ccc"
	}
	return nil
}

// GetTypeMeta ...
func (x *Foo) GetTypeMeta() *shared.TypeMeta {
	return &x.TypeMeta
}
