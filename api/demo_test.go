package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"k8s.io/kubeadm/api/scheme"
	"k8s.io/kubeadm/api/v1beta1"
	"k8s.io/kubeadm/api/v1beta2"
	"k8s.io/kubeadm/api/v1beta3"
)

var input = []byte(`
{
	"apiVersion": "kubeadm.k8s.io/v1beta2",
	"kind": "Foo",
	"a": "",
	"b": "bar"
}
---
{
	"apiVersion": "kubeadm.k8s.io/v1beta2",
	"kind": "Bar",
	"a": "a"
}
`)

func TestDemo(t *testing.T) {
	docs := bytes.Split(input, []byte("---\n"))
	v1beta1Foo := &v1beta1.Foo{}
	v1beta2Foo := &v1beta2.Foo{}
	v1beta3Foo := &v1beta3.Foo{}

	foundGroupVersionMap := map[string]string{}

	for _, doc := range docs {
		apiVersion, kind, err := scheme.ReadAPIVersionKindFromJSON(doc)
		if err != nil {
			t.Fatal(err)
		}
		group, version, err := scheme.ParseAPIVersion(apiVersion)
		if err != nil {
			t.Fatal(err)
		}

		ver, found := foundGroupVersionMap[group]
		if found && ver != version {
			t.Fatalf("mixed versions %s and %s for group %s", ver, version, group)
		}
		foundGroupVersionMap[group] = version

		if !scheme.IsKnownGroupVersion(group, version) {
			t.Fatal("unknown group / version")
		}
		if !scheme.IsKnownVersionKind(version, kind) {
			t.Fatalf("unknown version / kind")
		}
		t.Log("known", group, version, kind)

		switch version {
		case scheme.V1Beta1:
			switch kind {
			case scheme.V1Beta1Foo:
				if err := json.Unmarshal(doc, v1beta1Foo); err != nil {
					t.Fatal("unmarshal err", err)
				}
				v1beta1.DefaultFoo(v1beta1Foo)
				if err := v1beta1.ValidateFoo(v1beta1Foo); err != nil {
					t.Fatal("validate err", err)
				}
				v1beta2Foo, err = v1beta2.ConvertFoo(v1beta1Foo)
				if err != nil {
					t.Fatal("convert err", err)
				}
				v1beta3Foo, err = v1beta3.ConvertFoo(v1beta2Foo)
				if err != nil {
					t.Fatal("convert err", err)
				}
			}
		case scheme.V1Beta2:
			switch kind {
			case scheme.V1Beta2Foo:
				if err := json.Unmarshal(doc, v1beta2Foo); err != nil {
					t.Fatal("unmarshal err", err)
				}
				v1beta2.DefaultFoo(v1beta2Foo)
				if err := v1beta2.ValidateFoo(v1beta2Foo); err != nil {
					t.Fatal("validate err", err)
				}
				v1beta3Foo, err = v1beta3.ConvertFoo(v1beta2Foo)
				if err != nil {
					t.Fatal("convert err", err)
				}
			}
		case scheme.V1Beta3:
			switch kind {
			case scheme.V1Beta3Foo:
				if err := json.Unmarshal(doc, v1beta3Foo); err != nil {
					t.Fatal("unmarshal err", err)
				}
				v1beta3.DefaultFoo(v1beta3Foo)
			}
		}
	}
	t.Logf("%#v", v1beta3Foo)
}
