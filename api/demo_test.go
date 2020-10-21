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
`)

func TestDemo(t *testing.T) {
	cv := pkg.NewConverter(scheme.Group, scheme.VersionKinds)
	cv.SetUnmarshalFunc(json.Unmarshal)
	cv.SetMarshalFunc(json.Marshal)

	typemeta, err := cv.GetTypeMetaFromBytes(input)
	if err != nil {
		panic(err.Error())
	}
	obj, err := cv.GetObjectFromBytes(typemeta, input)
	if err != nil {
		panic(err.Error())
	}
	t.Logf("\n1--------%#v\n", obj)

	if err := obj.Default(); err != nil {
		panic(err.Error())
	}
	if err := obj.Validate(); err != nil {
		panic(err.Error())
	}

	old, err := cv.Marshal(obj)
	if err != nil {
		panic("marshal" + err.Error())
	}

	obj, err = cv.ConvertTo(obj, "v1beta1")
	if err != nil {
		panic(err.Error())
	}
	t.Logf("\n2-------- %#v\n", obj)

	data, err := cv.Marshal(obj)
	if err != nil {
		panic("marshal" + err.Error())
	}
	t.Logf("\n3-------- %s\n", data)

	obj, err = cv.ConvertTo(obj, "v1beta2")
	if err != nil {
		panic(err.Error())
	}
	t.Logf("\n4-------- %#v\n", obj)

	new, err := cv.Marshal(obj)
	if err != nil {
		panic("marshal" + err.Error())
	}

	if !bytes.Equal(old, new) {
		panic("could not roundtrip")
	}

	t.Logf("\n5-------- %s\n", new)
}
