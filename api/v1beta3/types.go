package v1beta3

import "k8s.io/kubeadm/api/shared"

// Zed ...
type Zed struct {
	shared.TypeMeta `json:",inline"`
	// A ...
	A string `json:"a,omitempty"`
	// B ...
	B string `json:"b,omitempty"`
	// C ...
	C string `json:"c,omitempty"`
}

var _ shared.Kind = (*Zed)(nil)

// Version ...
func (*Zed) Version() string {
	return "v1beta3"
}

// Name ...
func (*Zed) Name() string {
	return "Zed"
}

// ConvertUp ...
func (*Zed) ConvertUp(cv *shared.Converter, in shared.Kind) (shared.Kind, error) {
	return in, nil
}

// ConvertDown ...
func (*Zed) ConvertDown(cv *shared.Converter, in shared.Kind) (shared.Kind, error) {
	return in, nil
}

// ConvertUpName ...
func (*Zed) ConvertUpName() string {
	return "Zed"
}

// ConvertDownName ...
func (*Zed) ConvertDownName() string {
	return "Zed"
}

// Validate ...
func (*Zed) Validate(in shared.Kind) error {
	return nil
}

// Default ...
func (*Zed) Default(in shared.Kind) {
	return
}
