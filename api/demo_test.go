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

	"k8s.io/kubeadm/api/kubeadm"
	"k8s.io/kubeadm/api/kubeadm/groups"
	"k8s.io/kubeadm/api/kubeadm/v1beta2"
	"k8s.io/kubeadm/api/pkg"
)

var input = []byte(`
{
	"apiVersion": "kubeadm.k8s.io/v1beta2",
	"kind": "InitConfiguration",
	"certificateKey": "foo",
	"nodeRegistration": {
		"ignorePreflightErrors": ["bar"],
		"name": "foo",
		"criSocket": "` + v1beta2.DefaultURLScheme + `://foo"
	},
	"localAPIEndpoint": {
		"advertiseAddress": "127.0.0.1"
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
		"ignorePreflightErrors": ["bar"],
		"name": "foo",
		"criSocket": "` + v1beta2.DefaultURLScheme + `://foo"
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
	cv, err := pkg.NewConverter(kubeadm.Groups)
	if err != nil {
		t.Fatal(err)
	}
	cv.SetUnmarshalFunc(json.Unmarshal)
	cv.SetMarshalFunc(json.Marshal)

	docs, err := pkg.SplitDocuments(input)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("len of docs:", len(docs))

	for _, doc := range docs {

		typemeta, err := cv.TypeMetaFromBytes(doc)
		if err != nil {
			t.Fatal(err)
		}
		obj, err := cv.GetObjectFromBytes(typemeta, doc)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("\n1--------%#v\n", obj)

		if err := obj.Default(); err != nil {
			t.Fatal(err)
		}
		if err := obj.Validate(); err != nil {
			t.Fatal(err)
		}

		old, err := cv.Marshal(obj)
		if err != nil {
			t.Fatal("marshal", err)
		}

		spec := &pkg.ConvertSpec{Kinds: []pkg.Kind{obj}}
		spec, err = cv.ConvertTo(spec, groups.GroupKubeadm, "v1beta1")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("\n2-------- %#v\n", spec)

		data, err := cv.Marshal(spec.Kinds[0])
		if err != nil {
			t.Fatal("marshal", err)
		}
		t.Logf("\n3-------- %s\n", data)

		spec, err = cv.ConvertTo(spec, groups.GroupKubeadm, "v1beta2")
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("\n4-------- %#v\n", spec)

		new, err := cv.Marshal(spec.Kinds[0])
		if err != nil {
			t.Fatal("marshal", err)
		}

		if !bytes.Equal(old, new) {
			t.Fatal("could not roundtrip")
		}

		t.Logf("\n5-------- %s\n", new)
	}
}
