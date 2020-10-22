/*
Copyright 2017 The Kubernetes Authors.

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

package v1beta1

// FORK; there are failing tests here

import (
	// "io/ioutil" FORK;
	// "os" FORK;
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestValidateToken(t *testing.T) {
	var tests = []struct {
		token    string
		expected bool
	}{
		{"772ef5.6b6baab1d4a0a171", true},
		{".6b6baab1d4a0a171", false},
		{"772ef5.", false},
		{"abcdef.1234567890123456@foobar", false},
	}
	for _, rt := range tests {
		err := ValidateToken(rt.token, nil).ToAggregate()
		if (err == nil) != rt.expected {
			t.Errorf(
				"failed ValidateToken:\n\ttoken: %q\n\t  expected: %t, got: %t",
				rt.token,
				rt.expected,
				(err == nil),
			)
		}
	}
}

func TestValidateValidateTokenUsages(t *testing.T) {
	var tests = []struct {
		u        []string
		f        *field.Path
		expected bool
	}{
		{[]string{}, nil, true},                            // supported (no usages)
		{[]string{"signing", "authentication"}, nil, true}, // supported
		{[]string{"something else"}, nil, false},           // usage not supported
	}
	for _, rt := range tests {
		actual := ValidateTokenUsages(rt.u, rt.f)
		if (len(actual) == 0) != rt.expected {
			t.Errorf(
				"failed ValidateTokenUsages:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				(len(actual) == 0),
			)
		}
	}
}

func TestValidateTokenGroups(t *testing.T) {
	var tests = []struct {
		u        []string
		g        []string
		f        *field.Path
		expected bool
	}{
		{[]string{"some usage"}, []string{"some group"}, nil, false},                       // groups doesn't makes sense if usage authentication
		{[]string{"authentication"}, []string{"some group"}, nil, false},                   // group not supported
		{[]string{"authentication"}, []string{"system:bootstrappers:anygroup"}, nil, true}, // supported
	}
	for _, rt := range tests {
		actual := ValidateTokenGroups(rt.u, rt.g, rt.f)
		if (len(actual) == 0) != rt.expected {
			t.Errorf(
				"failed ValidateTokenGroups:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				(len(actual) == 0),
			)
		}
	}
}

func TestValidateNodeRegistrationOptions(t *testing.T) {
	var tests = []struct {
		nodeName       string
		expectedErrors bool
	}{
		{"", true},                  // node name can't be empty
		{"INVALID-NODENAME", true},  // Upper cases is invalid
		{"invalid-nodename-", true}, // Can't have trailing dashes
		{"invalid-node?name", true}, // Unsupported characters
		{"valid-nodename", false},   // supported
		// test cases for criSocket are covered in TestValidateSocketPath
	}
	for _, rt := range tests {
		nro := NodeRegistrationOptions{Name: rt.nodeName, CRISocket: "/some/path"}
		actual := ValidateNodeRegistrationOptions(&nro, field.NewPath("nodeRegistration"))
		actualErrors := len(actual) > 0
		if actualErrors != rt.expectedErrors {
			t.Errorf(
				"failed ValidateNodeRegistrationOptions: value: %v\n\texpected: %t\n\t  actual: %t",
				nro,
				rt.expectedErrors,
				actualErrors,
			)
		}
	}
}

func TestValidateCertSANs(t *testing.T) {
	var tests = []struct {
		sans     []string
		expected bool
	}{
		{[]string{}, true},                                                     // ok if not provided
		{[]string{"1,2,,3"}, false},                                            // not a DNS label or IP
		{[]string{"my-hostname", "???&?.garbage"}, false},                      // not valid
		{[]string{"my-hostname", "my.subdomain", "1.2.3.4"}, true},             // supported
		{[]string{"my-hostname2", "my.other.subdomain", "10.0.0.10"}, true},    // supported
		{[]string{"my-hostname", "my.subdomain", "2001:db8::4"}, true},         // supported
		{[]string{"my-hostname2", "my.other.subdomain", "2001:db8::10"}, true}, // supported
		{[]string{"*.my-hostname2", "*.my.other.subdomain"}, true},             // supported Wildcard DNS label
		{[]string{"**.my-hostname2", "my.other.subdomain"}, false},             // not a Wildcard DNS label
		{[]string{"*.*.my-hostname2", "my.other.subdomain"}, false},            // not a Wildcard DNS label
		{[]string{"a.*.my-hostname2", "my.other.subdomain"}, false},            // not a Wildcard DNS label
		{[]string{"*", "my.other.subdomain", "2001:db8::10"}, false},           // not a Wildcard DNS label
	}
	for _, rt := range tests {
		actual := ValidateCertSANs(rt.sans, nil)
		if (len(actual) == 0) != rt.expected {
			t.Errorf(
				"failed ValidateCertSANs:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				(len(actual) == 0),
			)
		}
	}
}

func TestValidateIPFromString(t *testing.T) {
	var tests = []struct {
		name     string
		ip       string
		expected bool
	}{
		{"invalid missing address", "", false},
		{"invalid missing decimal points in IPv4 address", "1234", false},
		{"invalid incomplete IPv4 address", "1.2", false},
		{"invalid IPv4 CIDR provided instead of IPv4 address", "1.2.3.4/16", false},
		{"valid IPv4 address", "1.2.3.4", true},
		{"valid IPv6 address", "2001:db8::1", true},
		{"invalid IPv6 CIDR provided instead of IPv6 address", "2001:db8::1/64", false},
		{"invalid hex character in IPv6 address", "2001:xb8::", false},
		{"invalid use of colons in IPv6 address", "2001::db8::", false},
	}
	for _, rt := range tests {
		actual := ValidateIPFromString(rt.ip, nil)
		if (len(actual) == 0) != rt.expected {
			t.Errorf(
				"%s test case failed:\n\texpected: %t\n\t  actual: %t",
				rt.name,
				rt.expected,
				(len(actual) == 0),
			)
		}
	}
}

func TestValidateIPNetFromString(t *testing.T) {
	var tests = []struct {
		name           string
		subnet         string
		minaddrs       int64
		checkDualStack bool
		expected       bool
	}{
		{"invalid missing CIDR", "", 0, false, false},
		{"invalid  CIDR", "a", 0, false, false},
		{"invalid CIDR missing decimal points in IPv4 address and / mask", "1234", 0, false, false},
		{"invalid CIDR use of letters instead of numbers and / mask", "abc", 0, false, false},
		{"invalid IPv4 address provided instead of CIDR representation", "1.2.3.4", 0, false, false},
		{"invalid IPv6 address provided instead of CIDR representation", "2001:db8::1", 0, false, false},
		{"invalid multiple CIDR provided in a single stack cluster", "2001:db8::1/64,1.2.3.4/24", 0, false, false},
		{"invalid multiple CIDR provided in a single stack cluster and one invalid subnet", "2001:db8::1/64,a", 0, false, false},
		{"valid, but IPv4 CIDR too small. At least 10 addresses needed", "10.0.0.16/29", 10, false, false},
		{"valid, but IPv6 CIDR too small. At least 10 addresses needed", "2001:db8::/125", 10, false, false},
		{"valid IPv4 CIDR", "10.0.0.16/12", 10, false, true},
		{"valid IPv6 CIDR", "2001:db8::/98", 10, false, true},
		// dual-stack:
		{"invalid missing CIDR", "", 0, true, false},
		{"valid dual-stack enabled but only an IPv4 CIDR specified", "10.0.0.16/12", 10, true, true},
		{"valid dual-stack enabled but only an IPv6 CIDR specified", "2001:db8::/98", 10, true, true},
		{"invalid IPv4 address provided instead of CIDR representation", "1.2.3.4,2001:db8::/98", 0, true, false},
		{"invalid IPv6 address provided instead of CIDR representation", "2001:db8::1,10.0.0.16/12", 0, true, false},
		{"valid, but IPv4 CIDR too small. At least 10 addresses needed", "10.0.0.16/29,2001:db8::/98", 10, true, false},
		{"valid, but IPv6 CIDR too small. At least 10 addresses needed", "10.0.0.16/12,2001:db8::/125", 10, true, false},
		{"valid, but only IPv4 family addresses specified. IPv6 CIDR is necessary.", "10.0.0.16/12,192.168.0.0/16", 10, true, false},
		{"valid, but only IPv6 family addresses specified. IPv4 CIDR is necessary.", "2001:db8::/98,2005:db8::/98", 10, true, false},
		{"valid IPv4 and IPv6 CIDR", "10.0.0.16/12,2001:db8::/98", 10, true, true},
		{"valid IPv6 and IPv4 CIDR", "10.0.0.16/12,2001:db8::/98", 10, true, true},
		{"invalid IPv6 and IPv4 CIDR with more than 2 subnets", "10.0.0.16/12,2001:db8::/98,192.168.0.0/16", 10, true, false},
		{"invalid IPv6 and IPv4 CIDR with more than 2 subnets", "10.0.0.16/12,2001:db8::/98,192.168.0.0/16,a.b.c.d/24", 10, true, false},
	}
	for _, rt := range tests {
		actual := ValidateIPNetFromString(rt.subnet, rt.minaddrs, rt.checkDualStack, nil)
		if (len(actual) == 0) != rt.expected {
			t.Errorf(
				"%s test case failed :\n\texpected: %t\n\t  actual: %t\n\t  err(s): %v\n\t",
				rt.name,
				rt.expected,
				(len(actual) == 0),
				actual,
			)
		}
	}
}

func TestValidateHostPort(t *testing.T) {
	var tests = []struct {
		name     string
		s        string
		expected bool
	}{
		{
			name:     "Valid DNS address / port",
			s:        "cp.k8s.io:8081",
			expected: true,
		},
		{
			name:     "Valid DNS address",
			s:        "cp.k8s.io",
			expected: true,
		},
		{
			name:     "Valid IPv4 address / port",
			s:        "1.2.3.4:8081",
			expected: true,
		},
		{
			name:     "Valid IPv4 address",
			s:        "1.2.3.4",
			expected: true,
		},
		{
			name:     "Valid IPv6 address / port",
			s:        "[2001:db7::1]:8081",
			expected: true,
		},
		{
			name:     "Valid IPv6 address",
			s:        "2001:db7::1",
			expected: true,
		},
		{
			name:     "Invalid IPv4 address, but valid DNS",
			s:        "1.2.34",
			expected: true,
		},
		{
			name:     "Invalid DNS",
			s:        "a.B.c.d.e",
			expected: false,
		},
		{
			name:     "Invalid IPv6 address",
			s:        "2001:db7:1",
			expected: false,
		},
		{
			name:     "Invalid BindPort",
			s:        "1.2.3.4:0",
			expected: false,
		},
	}
	for _, rt := range tests {
		actual := ValidateHostPort(rt.s, nil)
		if (len(actual) == 0) != rt.expected {
			t.Errorf(
				"%s test case failed:\n\texpected: %t\n\t  actual: %t",
				rt.name,
				rt.expected,
				(len(actual) == 0),
			)
		}
	}
}

func TestValidateAPIEndpoint(t *testing.T) {
	var tests = []struct {
		name     string
		s        *APIEndpoint
		expected bool
	}{
		{
			name: "Valid IPv4 address / port",
			s: &APIEndpoint{
				AdvertiseAddress: "4.5.6.7",
				BindPort:         6443,
			},
			expected: true,
		},
		{
			name: "Valid IPv6 address / port",
			s: &APIEndpoint{
				AdvertiseAddress: "2001:db7::2",
				BindPort:         6443,
			},
			expected: true,
		},
		{
			name: "Invalid IPv4 address",
			s: &APIEndpoint{
				AdvertiseAddress: "1.2.34",
				BindPort:         6443,
			},
			expected: false,
		},
		{
			name: "Invalid IPv6 address",
			s: &APIEndpoint{
				AdvertiseAddress: "2001:db7:1",
				BindPort:         6443,
			},
			expected: false,
		},
		{
			name: "Invalid BindPort",
			s: &APIEndpoint{
				AdvertiseAddress: "4.5.6.7",
				BindPort:         0,
			},
			expected: false,
		},
	}
	for _, rt := range tests {
		actual := ValidateAPIEndpoint(rt.s, nil)
		if (len(actual) == 0) != rt.expected {
			t.Errorf(
				"%s test case failed:\n\texpected: %t\n\t  actual: %t",
				rt.name,
				rt.expected,
				(len(actual) == 0),
			)
		}
	}
}

//TODO: Create a separated test for ValidateClusterConfiguration
func TestValidateInitConfiguration(t *testing.T) {
	nodename := "valid-nodename"
	var tests = []struct {
		name     string
		s        *InitConfiguration
		expected bool
	}{
		{"invalid missing InitConfiguration",
			&InitConfiguration{}, false},
		{"invalid missing token with IPv4 service subnet",
			&InitConfiguration{
				LocalAPIEndpoint: APIEndpoint{
					AdvertiseAddress: "1.2.3.4",
					BindPort:         6443,
				},
				ClusterConfiguration: ClusterConfiguration{
					Networking: Networking{
						ServiceSubnet: "10.96.0.1/12",
						DNSDomain:     "cluster.local",
					},
					CertificatesDir: "/some/cert/dir",
				},
				NodeRegistration: NodeRegistrationOptions{Name: nodename, CRISocket: "/some/path"},
			}, false},
		{"invalid missing token with IPv6 service subnet",
			&InitConfiguration{
				LocalAPIEndpoint: APIEndpoint{
					AdvertiseAddress: "1.2.3.4",
					BindPort:         6443,
				},
				ClusterConfiguration: ClusterConfiguration{
					Networking: Networking{
						ServiceSubnet: "2001:db8::1/98",
						DNSDomain:     "cluster.local",
					},
					CertificatesDir: "/some/cert/dir",
				},
				NodeRegistration: NodeRegistrationOptions{Name: nodename, CRISocket: "/some/path"},
			}, false},
		{"invalid missing node name",
			&InitConfiguration{
				LocalAPIEndpoint: APIEndpoint{
					AdvertiseAddress: "1.2.3.4",
					BindPort:         6443,
				},
				ClusterConfiguration: ClusterConfiguration{
					Networking: Networking{
						ServiceSubnet: "10.96.0.1/12",
						DNSDomain:     "cluster.local",
					},
					CertificatesDir: "/some/other/cert/dir",
				},
			}, false},
		{"valid InitConfiguration with incorrect IPv4 pod subnet",
			&InitConfiguration{
				LocalAPIEndpoint: APIEndpoint{
					AdvertiseAddress: "1.2.3.4",
					BindPort:         6443,
				},
				ClusterConfiguration: ClusterConfiguration{
					Networking: Networking{
						ServiceSubnet: "10.96.0.1/12",
						DNSDomain:     "cluster.local",
						PodSubnet:     "10.0.1.15",
					},
					CertificatesDir: "/some/other/cert/dir",
				},
				NodeRegistration: NodeRegistrationOptions{Name: nodename, CRISocket: "/some/path"},
			}, false},
		{"valid InitConfiguration with IPv4 service subnet",
			&InitConfiguration{
				LocalAPIEndpoint: APIEndpoint{
					AdvertiseAddress: "1.2.3.4",
					BindPort:         6443,
				},
				ClusterConfiguration: ClusterConfiguration{
					Etcd: Etcd{
						Local: &LocalEtcd{
							DataDir: "/some/path",
						},
					},
					Networking: Networking{
						ServiceSubnet: "10.96.0.1/12",
						DNSDomain:     "cluster.local",
						PodSubnet:     "10.0.1.15/16",
					},
					CertificatesDir: "/some/other/cert/dir",
				},
				NodeRegistration: NodeRegistrationOptions{Name: nodename, CRISocket: "/some/path"},
			}, true},
		{"valid InitConfiguration using IPv6 service subnet",
			&InitConfiguration{
				LocalAPIEndpoint: APIEndpoint{
					AdvertiseAddress: "1:2:3::4",
					BindPort:         3446,
				},
				ClusterConfiguration: ClusterConfiguration{
					Etcd: Etcd{
						Local: &LocalEtcd{
							DataDir: "/some/path",
						},
					},
					Networking: Networking{
						ServiceSubnet: "2001:db8::1/98",
						DNSDomain:     "cluster.local",
					},
					CertificatesDir: "/some/other/cert/dir",
				},
				NodeRegistration: NodeRegistrationOptions{Name: nodename, CRISocket: "/some/path"},
			}, true},
	}
	for _, rt := range tests {
		actual := ValidateInitConfiguration(rt.s)
		if (len(actual) == 0) != rt.expected {
			t.Errorf(
				"%s test case failed:\n\texpected: %t\n\t  actual: %t",
				rt.name,
				rt.expected,
				(len(actual) == 0),
			)
		}
	}
}

func TestValidateJoinConfiguration(t *testing.T) {
	var tests = []struct {
		s        *JoinConfiguration
		expected bool
	}{
		{&JoinConfiguration{}, false},
		{&JoinConfiguration{
			CACertPath: "/some/cert.crt",
			Discovery: Discovery{
				BootstrapToken: &BootstrapTokenDiscovery{
					Token: "abcdef.1234567890123456@foobar",
				},
				File: &FileDiscovery{
					KubeConfigPath: "foo",
				},
			},
		}, false},
		{&JoinConfiguration{ // Pass without JoinControlPlane
			CACertPath: "/some/cert.crt",
			Discovery: Discovery{
				BootstrapToken: &BootstrapTokenDiscovery{
					Token:             "abcdef.1234567890123456",
					APIServerEndpoint: "1.2.3.4:6443",
					CACertHashes:      []string{"aaaa"},
				},
				TLSBootstrapToken: "abcdef.1234567890123456",
			},
			NodeRegistration: NodeRegistrationOptions{
				Name:      "aaa",
				CRISocket: "/var/run/dockershim.sock",
			},
		}, true},
		{&JoinConfiguration{ // Pass with JoinControlPlane
			CACertPath: "/some/cert.crt",
			Discovery: Discovery{
				BootstrapToken: &BootstrapTokenDiscovery{
					Token:             "abcdef.1234567890123456",
					APIServerEndpoint: "1.2.3.4:6443",
					CACertHashes:      []string{"aaaa"},
				},
				TLSBootstrapToken: "abcdef.1234567890123456",
			},
			NodeRegistration: NodeRegistrationOptions{
				Name:      "aaa",
				CRISocket: "/var/run/dockershim.sock",
			},
			ControlPlane: &JoinControlPlane{
				LocalAPIEndpoint: APIEndpoint{
					AdvertiseAddress: "1.2.3.4",
					BindPort:         1234,
				},
			},
		}, true},
		{&JoinConfiguration{ // Fail JoinControlPlane.AdvertiseAddress validation
			CACertPath: "/some/cert.crt",
			Discovery: Discovery{
				BootstrapToken: &BootstrapTokenDiscovery{
					Token:             "abcdef.1234567890123456",
					APIServerEndpoint: "1.2.3.4:6443",
					CACertHashes:      []string{"aaaa"},
				},
				TLSBootstrapToken: "abcdef.1234567890123456",
			},
			NodeRegistration: NodeRegistrationOptions{
				Name:      "aaa",
				CRISocket: "/var/run/dockershim.sock",
			},
			ControlPlane: &JoinControlPlane{
				LocalAPIEndpoint: APIEndpoint{
					AdvertiseAddress: "aaa",
					BindPort:         1234,
				},
			},
		}, false},
		{&JoinConfiguration{ // Fail JoinControlPlane.BindPort validation
			CACertPath: "/some/cert.crt",
			Discovery: Discovery{
				BootstrapToken: &BootstrapTokenDiscovery{
					Token:             "abcdef.1234567890123456",
					APIServerEndpoint: "1.2.3.4:6443",
					CACertHashes:      []string{"aaaa"},
				},
				TLSBootstrapToken: "abcdef.1234567890123456",
			},
			NodeRegistration: NodeRegistrationOptions{
				Name:      "aaa",
				CRISocket: "/var/run/dockershim.sock",
			},
			ControlPlane: &JoinControlPlane{
				LocalAPIEndpoint: APIEndpoint{
					AdvertiseAddress: "1.2.3.4",
					BindPort:         -1,
				},
			},
		}, false},
	}
	for _, rt := range tests {
		actual := ValidateJoinConfiguration(rt.s)
		if (len(actual) == 0) != rt.expected {
			t.Errorf(
				"failed ValidateJoinConfiguration:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				(len(actual) == 0),
			)
		}
	}
}

// FORK; TestValidateFeatureGates is deleted

func TestValidateIgnorePreflightErrors(t *testing.T) {
	var tests = []struct {
		ignorePreflightErrorsFromCLI        []string
		ignorePreflightErrorsFromConfigFile []string
		expectedSet                         sets.String
		expectedError                       bool
	}{
		{ // empty lists in CLI and config file
			[]string{},
			[]string{},
			sets.NewString(),
			false,
		},
		{ // empty list in CLI only
			[]string{},
			[]string{"a"},
			sets.NewString("a"),
			false,
		},
		{ // empty list in config file only
			[]string{"a"},
			[]string{},
			sets.NewString("a"),
			false,
		},
		{ // no duplicates, no overlap
			[]string{"a", "b"},
			[]string{"c", "d"},
			sets.NewString("a", "b", "c", "d"),
			false,
		},
		{ // some duplicates, with some overlapping duplicates
			[]string{"a", "b", "a"},
			[]string{"c", "b"},
			sets.NewString("a", "b", "c"),
			false,
		},
		{ // non-duplicate, but 'all' present together with individual checks in CLI
			[]string{"a", "b", "all"},
			[]string{},
			sets.NewString(),
			true,
		},
		{ // empty list in CLI, but 'all' present in config file, which is forbidden
			[]string{},
			[]string{"all"},
			sets.NewString(),
			true,
		},
		{ // non-duplicate, but 'all' present in config file, which is forbidden
			[]string{"a", "b"},
			[]string{"all"},
			sets.NewString(),
			true,
		},
		{ // non-duplicate, but 'all' present in CLI, while values are in config file, which is forbidden
			[]string{"all"},
			[]string{"a", "b"},
			sets.NewString(),
			true,
		},
		{ // skip all checks
			[]string{"all"},
			[]string{},
			sets.NewString("all"),
			false,
		},
	}
	for _, rt := range tests {
		result, err := ValidateIgnorePreflightErrors(rt.ignorePreflightErrorsFromCLI, rt.ignorePreflightErrorsFromConfigFile)
		switch {
		case err != nil && !rt.expectedError:
			t.Errorf("ValidateIgnorePreflightErrors: unexpected error for input (%s, %s), error: %v", rt.ignorePreflightErrorsFromCLI, rt.ignorePreflightErrorsFromConfigFile, err)
		case err == nil && rt.expectedError:
			t.Errorf("ValidateIgnorePreflightErrors: expected error for input (%s, %s) but got: %v", rt.ignorePreflightErrorsFromCLI, rt.ignorePreflightErrorsFromConfigFile, result)
		case err == nil && !result.Equal(rt.expectedSet):
			t.Errorf("ValidateIgnorePreflightErrors: expected (%v) for input (%s, %s) but got: %v", rt.expectedSet, rt.ignorePreflightErrorsFromCLI, rt.ignorePreflightErrorsFromConfigFile, result)
		}
	}
}

func TestValidateDiscovery(t *testing.T) {
	var tests = []struct {
		name     string
		d        *Discovery
		expected bool
	}{
		{
			"invalid: .BootstrapToken and .File cannot both be set",
			&Discovery{
				BootstrapToken: &BootstrapTokenDiscovery{
					Token: "abcdef.1234567890123456",
				},
				File: &FileDiscovery{
					KubeConfigPath: "https://url/file.conf",
				},
			},
			false,
		},
		{
			"invalid: .BootstrapToken or .File must be set",
			&Discovery{},
			false,
		},
	}
	for _, rt := range tests {
		t.Run(rt.name, func(t *testing.T) {
			err := ValidateDiscovery(rt.d, nil).ToAggregate()
			if (err == nil) != rt.expected {
				t.Errorf(
					"test case failed: ValidateDiscovery:\n\texpected: %t\n\t  actual: %t",
					rt.expected,
					(err == nil),
				)
			}
		})
	}
}

func TestValidateDiscoveryBootstrapToken(t *testing.T) {
	var tests = []struct {
		name     string
		btd      *BootstrapTokenDiscovery
		expected bool
	}{
		{
			"invalid: .APIServerEndpoint not set",
			&BootstrapTokenDiscovery{
				Token: "abcdef.1234567890123456",
			},
			false,
		},
		{
			"invalid: using token-based discovery without .BootstrapToken.CACertHashes and .BootstrapToken.UnsafeSkipCAVerification",
			&BootstrapTokenDiscovery{
				Token:                    "abcdef.1234567890123456",
				APIServerEndpoint:        "192.168.122.100:6443",
				UnsafeSkipCAVerification: false,
			},
			false,
		},
		{
			"valid: using token-based discovery with .BootstrapToken.CACertHashes",
			&BootstrapTokenDiscovery{
				Token:                    "abcdef.1234567890123456",
				APIServerEndpoint:        "192.168.122.100:6443",
				CACertHashes:             []string{"sha256:7173b809ca12ec5dee4506cd86be934c4596dd234ee82c0662eac04a8c2c71dc"},
				UnsafeSkipCAVerification: false,
			},
			true,
		},
		{
			"valid: using token-based discovery with .BootstrapToken.CACertHashe but skip ca verification",
			&BootstrapTokenDiscovery{
				Token:                    "abcdef.1234567890123456",
				APIServerEndpoint:        "192.168.122.100:6443",
				CACertHashes:             []string{"sha256:7173b809ca12ec5dee4506cd86be934c4596dd234ee82c0662eac04a8c2c71dc"},
				UnsafeSkipCAVerification: true,
			},
			true,
		},
	}
	for _, rt := range tests {
		t.Run(rt.name, func(t *testing.T) {
			err := ValidateDiscoveryBootstrapToken(rt.btd, nil).ToAggregate()
			if (err == nil) != rt.expected {
				t.Errorf(
					"test case failed: ValidateDiscoveryBootstrapToken:\n\texpected: %t\n\t  actual: %t",
					rt.expected,
					(err == nil),
				)
			}
		})
	}
}

func TestValidateDiscoveryTokenAPIServer(t *testing.T) {
	var tests = []struct {
		apiServerEndpoint string
		expected          bool
	}{
		{
			"192.168.122.100",
			false,
		},
		{
			"192.168.122.100:6443",
			true,
		},
	}
	for _, rt := range tests {
		actual := ValidateDiscoveryTokenAPIServer(rt.apiServerEndpoint, nil)
		if (len(actual) == 0) != rt.expected {
			t.Errorf(
				"failed ValidateDiscoveryTokenAPIServer:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				(len(actual) == 0),
			)
		}
	}
}

// FORK disabled; see ValidateDiscoveryKubeConfigPath
// func TestValidateDiscoveryKubeConfigPath(t *testing.T) {
// 	tmpfile, err := ioutil.TempFile("", "test_discovery_file")
// 	if err != nil {
// 		t.Errorf("Error creating temporary file: %v", err)
// 	}
// 	defer os.Remove(tmpfile.Name())

// 	var tests = []struct {
// 		s        string
// 		expected bool
// 	}{
// 		{"foo", false},
// 		{"/foo/bar/file_which_i_believe_not_existing.conf", false},
// 		{tmpfile.Name(), true},
// 		{"http://[::1]a", false},
// 		{"http://url/file.conf", false},
// 		{"https://u r l/file.conf", false},
// 		{"https://url/file.conf", true},
// 	}
// 	for i, rt := range tests {
// 		actual := ValidateDiscoveryKubeConfigPath(rt.s, nil)
// 		if (len(actual) == 0) != rt.expected {
// 			t.Errorf(
// 				"%d: failed ValidateDiscoveryKubeConfigPath:\n\texpected: %t\n\t  actual: %t, %v",
// 				i,
// 				rt.expected,
// 				(len(actual) == 0),
// 				actual,
// 			)
// 		}
// 	}
// }

func TestValidateSocketPath(t *testing.T) {
	var tests = []struct {
		name           string
		criSocket      string
		expectedErrors bool
	}{
		{name: "valid path", criSocket: "/some/path", expectedErrors: false},
		{name: "valid socket url", criSocket: DefaultURLScheme + "://" + "/some/path", expectedErrors: false},
		{name: "unsupported url scheme", criSocket: "bla:///some/path", expectedErrors: true},
		{name: "unparseable url", criSocket: ":::", expectedErrors: true},
		{name: "invalid CRISocket (path is not absolute)", criSocket: "some/path", expectedErrors: true},
		{name: "empty CRISocket (path is not absolute)", criSocket: "", expectedErrors: true},
	}
	for _, tc := range tests {
		actual := ValidateSocketPath(tc.criSocket, field.NewPath("criSocket"))
		actualErrors := len(actual) > 0
		if actualErrors != tc.expectedErrors {
			t.Errorf("error: socket path: %q\n\texpected: %t\n\t  actual: %t", tc.criSocket, tc.expectedErrors, actualErrors)
		}
	}
}

func TestValidateURLs(t *testing.T) {
	var tests = []struct {
		name           string
		urls           []string
		requireHTTPS   bool
		expectedErrors bool
	}{
		{
			name:           "valid urls (https not required)",
			urls:           []string{"http://example.com", "https://example.org"},
			requireHTTPS:   false,
			expectedErrors: false,
		},
		{
			name:           "valid urls (https required)",
			urls:           []string{"https://example.com", "https://example.org"},
			requireHTTPS:   true,
			expectedErrors: false,
		},
		{
			name:           "invalid url (https required)",
			urls:           []string{"http://example.com", "https://example.org"},
			requireHTTPS:   true,
			expectedErrors: true,
		},
		{
			name:           "URL parse error",
			urls:           []string{"::://example.com"},
			requireHTTPS:   false,
			expectedErrors: true,
		},
		{
			name:           "URL without scheme",
			urls:           []string{"example.com"},
			requireHTTPS:   false,
			expectedErrors: true,
		},
	}
	for _, tc := range tests {
		actual := ValidateURLs(tc.urls, tc.requireHTTPS, nil)
		actualErrors := len(actual) > 0
		if actualErrors != tc.expectedErrors {
			t.Errorf("error:\n\texpected: %t\n\t  actual: %t", tc.expectedErrors, actualErrors)
		}
	}
}

func TestValidateEtcd(t *testing.T) {
	var tests = []struct {
		name           string
		etcd           *Etcd
		expectedErrors bool
	}{
		{
			name:           "either .Etcd.Local or .Etcd.External is required",
			etcd:           &Etcd{},
			expectedErrors: true,
		},
		{
			name: ".Etcd.Local and .Etcd.External are mutually exclusive",
			etcd: &Etcd{
				Local: &LocalEtcd{
					DataDir: "/some/path",
				},
				External: &ExternalEtcd{
					Endpoints: []string{"10.100.0.1:2379", "10.100.0.2:2379"},
				},
			},
			expectedErrors: true,
		},
		{
			name: "either both or none of .Etcd.External.CertFile and .Etcd.External.KeyFile must be set",
			etcd: &Etcd{
				External: &ExternalEtcd{
					Endpoints: []string{"https://external.etcd1:2379", "https://external.etcd2:2379"},
					CertFile:  "/some/file.crt",
				},
			},
			expectedErrors: true,
		},
		{
			name: "setting .Etcd.External.CertFile and .Etcd.External.KeyFile requires .Etcd.External.CAFile",
			etcd: &Etcd{
				External: &ExternalEtcd{
					Endpoints: []string{"https://external.etcd1:2379", "https://external.etcd2:2379"},
					CertFile:  "/some/file.crt",
					KeyFile:   "/some/file.key",
				},
			},
			expectedErrors: true,
		},
		{
			name: "valid external etcd",
			etcd: &Etcd{
				External: &ExternalEtcd{
					Endpoints: []string{"https://external.etcd1:2379", "https://external.etcd2:2379"},
					CertFile:  "/etcd.crt",
					KeyFile:   "/etcd.key",
					CAFile:    "/etcd-ca.crt",
				},
			},
			expectedErrors: false,
		},
		{
			name: "valid external etcd (no TLS)",
			etcd: &Etcd{
				External: &ExternalEtcd{
					Endpoints: []string{"http://10.100.0.1:2379", "http://10.100.0.2:2379"},
				},
			},
			expectedErrors: false,
		},
	}

	for _, tc := range tests {
		actual := ValidateEtcd(tc.etcd, field.NewPath("etcd"))
		actualErrors := len(actual) > 0
		if actualErrors != tc.expectedErrors {
			t.Errorf("Error: \n\texpected: %t\n\t  actual: %t",
				tc.expectedErrors,
				actualErrors,
			)
		}
	}
}
