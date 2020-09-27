package v1beta2

import "k8s.io/kubeadm/api/v1beta1"

// ConvertFoo ...
func ConvertFoo(in *v1beta1.Foo) (*Foo, error) {
	out := &Foo{}
	if len(in.A) != 0 {
		out.A = in.A
	}
	if len(in.B) != 0 {
		out.B = in.B
	}
	return out, nil
}
