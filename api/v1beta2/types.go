package v1beta2

import (
	"k8s.io/kubeadm/api/external/metav1"
	"k8s.io/kubeadm/api/shared"
	"k8s.io/kubeadm/api/v1beta1"
)

// Version ...
const Version = "v1beta2"

// Bar ...
type Bar struct {
	metav1.TypeMeta `json:",inline"`
	// A ...
	A string `json:"a,omitempty"`
	// B ...
	B string `json:"b,omitempty"`
}

var _ shared.Kind = (*Bar)(nil)

// Version ...
func (*Bar) Version() string {
	return Version
}

// Name ...
func (*Bar) Name() string {
	return "Bar"
}

// ConvertUp ...
func (*Bar) ConvertUp(cv *shared.Converter, in shared.Kind) (shared.Kind, error) {
	cv.AddToCache(in)
	obj, _ := in.(*v1beta1.Foo)
	out := &Bar{}
	cv.SetTypeMeta(out)
	out.A = obj.A
	out.B = obj.B
	return out, nil
}

// ConvertDown ...
func (*Bar) ConvertDown(cv *shared.Converter, in shared.Kind) (shared.Kind, error) {
	oldKind := cv.GetFromCache(&v1beta1.Foo{})
	foo := oldKind.(*v1beta1.Foo)
	obj, _ := in.(*Bar)
	foo.A = obj.A
	foo.B = obj.B
	return foo, nil
}

// ConvertUpName ...
func (*Bar) ConvertUpName() string {
	return "Foo"
}

// ConvertDownName ...
func (*Bar) ConvertDownName() string {
	return "Bar"
}

// Validate ...
func (x *Bar) Validate() error {
	return nil
}

// Default ...
func (x *Bar) Default() error {
	return nil
}

// GetTypeMeta ...
func (x *Bar) GetTypeMeta() *metav1.TypeMeta {
	return &x.TypeMeta
}
