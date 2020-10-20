/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta2

import (
	"net/url"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	bootstrapapi "k8s.io/cluster-bootstrap/token/api"
)

const (
	// DefaultServiceDNSDomain defines default cluster-internal domain name for Services and Pods
	DefaultServiceDNSDomain = "cluster.local"
	// DefaultServicesSubnet defines default service subnet range
	DefaultServicesSubnet = "10.96.0.0/12"
	// DefaultClusterDNSIP defines default DNS IP
	DefaultClusterDNSIP = "10.96.0.10"
	// DefaultKubernetesVersion defines default kubernetes version
	DefaultKubernetesVersion = "stable-1"
	// DefaultAPIBindPort defines default API port
	DefaultAPIBindPort = 6443
	// DefaultImageRepository defines default image registry
	DefaultImageRepository = "k8s.gcr.io"
	// DefaultClusterName defines the default cluster name
	DefaultClusterName = "kubernetes"

	// DefaultProxyBindAddressv4 is the default bind address when the advertise address is v4
	DefaultProxyBindAddressv4 = "0.0.0.0"
	// DefaultProxyBindAddressv6 is the default bind address when the advertise address is v6
	DefaultProxyBindAddressv6 = "::"
	// DefaultDiscoveryTimeout specifies the default discovery timeout for kubeadm (used unless one is specified in the JoinConfiguration)
	DefaultDiscoveryTimeout = 5 * time.Minute

	// COPIED FROM CONSTANTS

	// DefaultControlPlaneTimeout ...
	DefaultControlPlaneTimeout = 4 * time.Minute

	// NodeBootstrapTokenAuthGroup specifies which group a Node Bootstrap Token should be authenticated in
	NodeBootstrapTokenAuthGroup = "system:bootstrappers:kubeadm:default-node-token"

	// DefaultTokenDuration specifies the default amount of time that a bootstrap token will be valid
	// Default behaviour is 24 hours
	DefaultTokenDuration = 24 * time.Hour
)

var (
	// DefaultTokenUsages specifies the default functions a token will get
	DefaultTokenUsages = bootstrapapi.KnownTokenUsages

	// DefaultTokenGroups specifies the default groups that this token will authenticate as when used for authentication
	DefaultTokenGroups = []string{NodeBootstrapTokenAuthGroup}
)

// Default ...
func (obj *InitConfiguration) Default() error {
	SetDefaultsBootstrapTokens(obj)
	SetDefaultsAPIEndpoint(&obj.LocalAPIEndpoint)
	return nil
}

// Default ...
func (obj *ClusterConfiguration) Default() error {
	if obj.KubernetesVersion == "" {
		obj.KubernetesVersion = DefaultKubernetesVersion
	}

	if obj.Networking.ServiceSubnet == "" {
		obj.Networking.ServiceSubnet = DefaultServicesSubnet
	}

	if obj.Networking.DNSDomain == "" {
		obj.Networking.DNSDomain = DefaultServiceDNSDomain
	}

	if obj.CertificatesDir == "" {
		obj.CertificatesDir = DefaultCertificatesDir
	}

	if obj.ImageRepository == "" {
		obj.ImageRepository = DefaultImageRepository
	}

	if obj.ClusterName == "" {
		obj.ClusterName = DefaultClusterName
	}

	SetDefaultsDNS(obj)
	SetDefaultsEtcd(obj)
	SetDefaultsAPIServer(&obj.APIServer)

	return nil
}

// Default ...
func (*ClusterStatus) Default() error {
	return nil
}

// Default ...
func (obj *JoinConfiguration) Default() error {
	if obj.CACertPath == "" {
		obj.CACertPath = DefaultCACertPath
	}

	SetDefaultsJoinControlPlane(obj.ControlPlane)
	SetDefaultsDiscovery(&obj.Discovery)

	return nil
}

// ------------------------------

// SetDefaultsAPIServer assigns default values for the API Server
func SetDefaultsAPIServer(obj *APIServer) {
	if obj.TimeoutForControlPlane == nil {
		obj.TimeoutForControlPlane = &metav1.Duration{
			Duration: DefaultControlPlaneTimeout,
		}
	}
}

// SetDefaultsDNS assigns default values for the DNS component
func SetDefaultsDNS(obj *ClusterConfiguration) {
	if obj.DNS.Type == "" {
		obj.DNS.Type = CoreDNS
	}
}

// SetDefaultsEtcd assigns default values for the proxy
func SetDefaultsEtcd(obj *ClusterConfiguration) {
	if obj.Etcd.External == nil && obj.Etcd.Local == nil {
		obj.Etcd.Local = &LocalEtcd{}
	}
	if obj.Etcd.Local != nil {
		if obj.Etcd.Local.DataDir == "" {
			obj.Etcd.Local.DataDir = DefaultEtcdDataDir
		}
	}
}

// SetDefaultsJoinControlPlane ...
func SetDefaultsJoinControlPlane(obj *JoinControlPlane) {
	if obj != nil {
		SetDefaultsAPIEndpoint(&obj.LocalAPIEndpoint)
	}
}

// SetDefaultsDiscovery assigns default values for the discovery process
func SetDefaultsDiscovery(obj *Discovery) {
	if len(obj.TLSBootstrapToken) == 0 && obj.BootstrapToken != nil {
		obj.TLSBootstrapToken = obj.BootstrapToken.Token
	}

	if obj.Timeout == nil {
		obj.Timeout = &metav1.Duration{
			Duration: DefaultDiscoveryTimeout,
		}
	}

	if obj.File != nil {
		SetDefaultsFileDiscovery(obj.File)
	}
}

// SetDefaultsFileDiscovery assigns default values for file based discovery
func SetDefaultsFileDiscovery(obj *FileDiscovery) {
	// Make sure file URL becomes path
	if len(obj.KubeConfigPath) != 0 {
		u, err := url.Parse(obj.KubeConfigPath)
		if err == nil && u.Scheme == "file" {
			obj.KubeConfigPath = u.Path
		}
	}
}

// SetDefaultsBootstrapTokens sets the defaults for the .BootstrapTokens field
// If the slice is empty, it's defaulted with one token. Otherwise it just loops
// through the slice and sets the defaults for the omitempty fields that are TTL,
// Usages and Groups. Token is NOT defaulted with a random one in the API defaulting
// layer, but set to a random value later at runtime if not set before.
func SetDefaultsBootstrapTokens(obj *InitConfiguration) {

	if obj.BootstrapTokens == nil || len(obj.BootstrapTokens) == 0 {
		obj.BootstrapTokens = []BootstrapToken{{}}
	}

	for i := range obj.BootstrapTokens {
		SetDefaultsBootstrapToken(&obj.BootstrapTokens[i])
	}
}

// SetDefaultsBootstrapToken sets the defaults for an individual Bootstrap Token
func SetDefaultsBootstrapToken(bt *BootstrapToken) {
	if bt.TTL == nil {
		bt.TTL = &metav1.Duration{
			Duration: DefaultTokenDuration,
		}
	}
	if len(bt.Usages) == 0 {
		bt.Usages = DefaultTokenUsages
	}

	if len(bt.Groups) == 0 {
		bt.Groups = DefaultTokenGroups
	}
}

// SetDefaultsAPIEndpoint sets the defaults for the API server instance deployed on a node.
func SetDefaultsAPIEndpoint(obj *APIEndpoint) {
	if obj.BindPort == 0 {
		obj.BindPort = DefaultAPIBindPort
	}
}
