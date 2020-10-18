package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/kubeadm/api/pkg"
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

var _ pkg.Kind = (*Bar)(nil)

// Version ...
func (*Bar) Version() string {
	return Version
}

// Name ...
func (*Bar) Name() string {
	return "Bar"
}

// ConvertUp ...
func (*Bar) ConvertUp(cv *pkg.Converter, in pkg.Kind) (pkg.Kind, error) {
	/*
		cv.AddToCache(in)
		obj, _ := in.(*v1beta1.Foo)
		out := &Bar{}
		cv.SetTypeMeta(out)
		out.A = obj.A
		out.B = obj.B
		return out, ni
	*/
	return nil, nil
}

// ConvertDown ...
func (*Bar) ConvertDown(cv *pkg.Converter, in pkg.Kind) (pkg.Kind, error) {
	/*
		oldKind := cv.GetFromCache(&v1beta1.Foo{})
		foo := oldKind.(*v1beta1.Foo)
		obj, _ := in.(*Bar)
		foo.A = obj.A
		foo.B = obj.B
		return foo, nil
	*/
	return nil, nil
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
