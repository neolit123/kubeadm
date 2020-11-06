/*
Copyright 2020 The Kubernetes Authors.

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

package shared

import (
	"net"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/util/validation"
	utilsnet "k8s.io/utils/net"
)

// ParsePort parses a string representing a TCP port.
// If the string is not a valid representation of a TCP port, ParsePort returns an error.
func ParsePort(port string) (int, error) {
	portInt, err := utilsnet.ParsePort(port, true)
	if err == nil && (1 <= portInt && portInt <= 65535) {
		return portInt, nil
	}

	return 0, errors.New("port must be a valid number between 1 and 65535, inclusive")
}

// ParseHostPort parses a network address of the form "host:port", "ipv4:port", "[ipv6]:port" into host and port;
// ":port" can be eventually omitted.
// If the string is not a valid representation of network address, ParseHostPort returns an error.
func ParseHostPort(hostport string) (string, string, error) {
	var host, port string
	var err error

	// try to split host and port
	if host, port, err = net.SplitHostPort(hostport); err != nil {
		// if SplitHostPort returns an error, the entire hostport is considered as host
		host = hostport
	}

	// if port is defined, parse and validate it
	if port != "" {
		if _, err := ParsePort(port); err != nil {
			return "", "", errors.Errorf("hostport %s: port %s must be a valid number between 1 and 65535, inclusive", hostport, port)
		}
	}

	// if host is a valid IP, returns it
	if ip := net.ParseIP(host); ip != nil {
		return host, port, nil
	}

	// if host is a validate RFC-1123 subdomain, returns it
	if errs := validation.IsDNS1123Subdomain(host); len(errs) == 0 {
		return host, port, nil
	}

	return "", "", errors.Errorf("hostport %s: host '%s' must be a valid IP address or a valid RFC-1123 DNS subdomain", hostport, host)
}
