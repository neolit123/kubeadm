package v1beta3

import (
	"k8s.io/kubeadm/api/shared"
	"k8s.io/kubeadm/api/v1beta2"
)

// Zed ...
type Zed struct {
	shared.TypeMeta `json:",inline"`
	// A ...
	A string `json:"a,omitempty"`
	// B ...
	B string `json:"b,omitempty"`
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
	cv.AddToCache("v1beta2.Bar", in)
	obj, _ := in.(*v1beta2.Bar)
	out := &Zed{}
	cv.SetTypeMeta(out)
	out.A = obj.A
	out.B = obj.B
	return out, nil
}

// ConvertDown ...
func (*Zed) ConvertDown(cv *shared.Converter, in shared.Kind) (shared.Kind, error) {
	obj := in.(*Zed)
	bar := &v1beta2.Bar{}
	bar.A = obj.A
	bar.B = obj.B
	return bar, nil
}

// ConvertUpName ...
func (*Zed) ConvertUpName() string {
	return "Bar"
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

// GetTypeMeta ...
func (x *Zed) GetTypeMeta() *shared.TypeMeta {
	return &x.TypeMeta
}
