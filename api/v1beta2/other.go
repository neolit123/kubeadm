package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Version ...
const Version = "v1beta2"

// GetTypeMeta ...
func (x *InitConfiguration) GetTypeMeta() *metav1.TypeMeta {
	return &x.TypeMeta
}

// Name ...
func (*InitConfiguration) Name() string {
	return "InitConfiguration"
}

// Version ...
func (*InitConfiguration) Version() string {
	return Version
}

// ---------------

// GetTypeMeta ...
func (x *ClusterConfiguration) GetTypeMeta() *metav1.TypeMeta {
	return &x.TypeMeta
}

// Name ...
func (*ClusterConfiguration) Name() string {
	return "ClusterConfiguration"
}

// Version ...
func (*ClusterConfiguration) Version() string {
	return Version
}

// ---------------

// GetTypeMeta ...
func (x *ClusterStatus) GetTypeMeta() *metav1.TypeMeta {
	return &x.TypeMeta
}

// Name ...
func (*ClusterStatus) Name() string {
	return "ClusterStatus"
}

// Version ...
func (*ClusterStatus) Version() string {
	return Version
}

// ---------------

// GetTypeMeta ...
func (x *JoinConfiguration) GetTypeMeta() *metav1.TypeMeta {
	return &x.TypeMeta
}

// Name ...
func (*JoinConfiguration) Name() string {
	return "JoinConfiguration"
}

// Version ...
func (*JoinConfiguration) Version() string {
	return Version
}
