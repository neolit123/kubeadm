package v1beta3

import (
	"k8s.io/kubeadm/api/shared"
	"k8s.io/kubeadm/api/v1beta2"
)

const Version = "v1beta3"

// Zed ...
type Zed struct {
	shared.TypeMeta `json:",inline"`
	// A ...
	A string `json:"a,omitempty"`
}

var _ shared.Kind = (*Zed)(nil)

// Version ...
func (*Zed) Version() string {
	return Version
}

// Name ...
func (*Zed) Name() string {
	return "Zed"
}

// ConvertUp ...
func (*Zed) ConvertUp(cv *shared.Converter, in shared.Kind) (shared.Kind, error) {
	cv.AddToCache(in)
	obj, _ := in.(*v1beta2.Bar)
	out := &Zed{}
	cv.SetTypeMeta(out)
	out.A = obj.A
	return out, nil
}

// ConvertDown ...
func (*Zed) ConvertDown(cv *shared.Converter, in shared.Kind) (shared.Kind, error) {
	oldKind := cv.GetFromCache(&v1beta2.Bar{})
	bar := oldKind.(*v1beta2.Bar)
	obj, _ := in.(*Zed)
	bar.A = obj.A
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
func (x *Zed) Validate() error {
	return nil
}

// Default ...
func (x *Zed) Default() error {
	return nil
}

// GetTypeMeta ...
func (x *Zed) GetTypeMeta() *shared.TypeMeta {
	return &x.TypeMeta
}
