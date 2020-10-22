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

package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"k8s.io/kubeadm/api/pkg"
	"k8s.io/kubeadm/api/scheme"
)

var input = []byte(`
{
	"apiVersion": "kubeadm.k8s.io/v1beta2",
	"kind": "InitConfiguration",
	"certificateKey": "foo",
	"nodeRegistration": {
		"ignorePreflightErrors": ["bar"]
	}
}
---
{
	"apiVersion": "kubeadm.k8s.io/v1beta2",
	"kind": "ClusterConfiguration",
	"kubernetesVersion": "v1.19.0"
}
---
{
	"apiVersion": "kubeadm.k8s.io/v1beta2",
	"kind": "JoinConfiguration",
	"nodeRegistration": {
		"ignorePreflightErrors": ["bar"]
	},
	"discovery": {
		"bootstrapToken": {
			"token": "abcdef.1234567890abcdef",
			"apiServerEndpoint": "1.2.3.4:6443",
			"unsafeSkipCAVerification": true
		}
	}
}
---
{
	"apiVersion": "kubeadm.k8s.io/v1beta2",
	"kind": "ClusterStatus",
	"apiEndpoints": {
		"foo": {
			"advertiseAddress": "1.2.3.4",
			"bindPort": 6443
		}
	}
}
`)

func TestDemo(t *testing.T) {
	cv := pkg.NewConverter(scheme.Group, scheme.VersionKinds)
	cv.SetUnmarshalFunc(json.Unmarshal)
	cv.SetMarshalFunc(json.Marshal)

	docs, err := cv.SplitDocuments(input)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log("len of docs:", len(docs))

	for _, doc := range docs {

		typemeta, err := cv.GetTypeMetaFromBytes(doc)
		if err != nil {
			t.Fatal(err.Error())
		}
		obj, err := cv.GetObjectFromBytes(typemeta, doc)
		if err != nil {
			t.Fatal(err.Error())
		}
		t.Logf("\n1--------%#v\n", obj)

		if err := obj.Default(); err != nil {
			t.Fatal(err.Error())
		}
		if err := obj.Validate(); err != nil {
			t.Fatal(err.Error())
		}

		old, err := cv.Marshal(obj)
		if err != nil {
			t.Fatal("marshal" + err.Error())
		}

		obj, err = cv.ConvertTo(obj, "v1beta1")
		if err != nil {
			t.Fatal(err.Error())
		}
		t.Logf("\n2-------- %#v\n", obj)

		data, err := cv.Marshal(obj)
		if err != nil {
			t.Fatal("marshal" + err.Error())
		}
		t.Logf("\n3-------- %s\n", data)

		obj, err = cv.ConvertTo(obj, "v1beta2")
		if err != nil {
			t.Fatal(err.Error())
		}
		t.Logf("\n4-------- %#v\n", obj)

		new, err := cv.Marshal(obj)
		if err != nil {
			t.Fatal("marshal" + err.Error())
		}

		if !bytes.Equal(old, new) {
			t.Fatal("could not roundtrip")
		}

		t.Logf("\n5-------- %s\n", new)
	}
}
