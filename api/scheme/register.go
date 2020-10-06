package scheme

import (
	"k8s.io/kubeadm/api/shared"
	"k8s.io/kubeadm/api/v1beta1"
	"k8s.io/kubeadm/api/v1beta2"
	"k8s.io/kubeadm/api/v1beta3"
)

const (
	// Group ...
	Group = "kubeadm.k8s.io"
)

// VersionKinds ...
var VersionKinds = []shared.VersionKinds{
	{
		Version: "v1beta1",
		Kinds: []shared.Kind{
			&v1beta1.Foo{},
		},
	},
	{
		Version: "v1beta2",
		Kinds: []shared.Kind{
			&v1beta2.Bar{},
		},
	},
	{
		Version: "v1beta3",
		Kinds: []shared.Kind{
			&v1beta3.Zed{},
		},
	},
}
