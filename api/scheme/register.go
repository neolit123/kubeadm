package scheme

import (
	"k8s.io/kubeadm/api/pkg"
	"k8s.io/kubeadm/api/v1beta1"
	"k8s.io/kubeadm/api/v1beta2"
)

const (
	// Group ...
	Group = "kubeadm.k8s.io"
)

// VersionKinds ...
var VersionKinds = []pkg.VersionKinds{
	{
		Version: v1beta1.Version,
		Kinds: []pkg.Kind{
			&v1beta1.InitConfiguration{},
			&v1beta1.ClusterConfiguration{},
			&v1beta1.ClusterStatus{},
			&v1beta1.JoinConfiguration{},
		},
	},
	{
		Version: v1beta2.Version,
		Kinds: []pkg.Kind{
			&v1beta2.Bar{},
		},
	},
}
