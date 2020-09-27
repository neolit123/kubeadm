package v1beta1

import "k8s.io/kubeadm/api/scheme"

// Foo ...
type Foo struct {
	scheme.TypeMeta `json:",inline"`
	// A ...
	A string `json:"a,omitempty"`
	// B ...
	B string `json:"b,omitempty"`
}
