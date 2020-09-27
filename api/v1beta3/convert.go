package v1beta3

import "k8s.io/kubeadm/api/v1beta2"

// ConvertFoo ...
func ConvertFoo(in *v1beta2.Foo) (*Foo, error) {
	out := &Foo{}
	if len(in.A) != 0 {
		out.A = in.A
	}
	if len(in.B) != 0 {
		out.B = in.B
	}
	if in.N != nil {
		out.N = &Bar{}
		if len(in.N.P) != 0 {
			out.N.P = in.N.P
		}
		if len(in.N.P) != 0 {
			out.N.P = in.N.P
		}
	}
	return out, nil
}
