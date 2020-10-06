package v1beta2

import "k8s.io/kubeadm/api/shared"

// Bar ...
type Bar struct {
	shared.TypeMeta `json:",inline"`
	// A ...
	A string `json:"a,omitempty"`
	// B ...
	B string `json:"b,omitempty"`
	// C ...
	C string `json:"c,omitempty"`
}

var _ shared.Kind = (*Bar)(nil)

// Version ...
func (*Bar) Version() string {
	return "v1beta2"
}

// Name ...
func (*Bar) Name() string {
	return "Bar"
}

// ConvertUp ...
func (*Bar) ConvertUp(cv *shared.Converter, in shared.Kind) (shared.Kind, error) {
	return in, nil
}

// ConvertDown ...
func (*Bar) ConvertDown(cv *shared.Converter, in shared.Kind) (shared.Kind, error) {
	return in, nil
}

// ConvertUpName ...
func (*Bar) ConvertUpName() string {
	return "Bar"
}

// ConvertDownName ...
func (*Bar) ConvertDownName() string {
	return "Bar"
}

// Validate ...
func (*Bar) Validate(in shared.Kind) error {
	return nil
}

// Default ...
func (*Bar) Default(in shared.Kind) {
	return
}
