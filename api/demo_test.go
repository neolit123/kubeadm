package main

import (
	"encoding/json"
	"testing"

	"k8s.io/kubeadm/api/pkg"
	"k8s.io/kubeadm/api/scheme"
)

var input = []byte(`
{
	"apiVersion": "kubeadm.k8s.io/v1beta1",
	"kind": "ClusterConfiguration",
	"kubernetesVersion": "foo"
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
	t.Logf("--------%#v", obj)

	if err := obj.Default(); err != nil {
		panic(err.Error())
	}
	if err := obj.Validate(); err != nil {
		panic(err.Error())
	}

	// obj, _ = cv.ConvertTo(obj, "v1beta3")
	// t.Logf("--------%#v", obj)

	// obj, _ = cv.ConvertTo(obj, "v1beta1")
	// t.Logf("--------%#v", obj)

	// obj, _ = cv.ConvertTo(obj, "v1beta2")
	// t.Logf("--------%#v", obj)

	// obj, err = obj.ConvertDown(cv, obj)
	// t.Logf("--------%#v", obj)

	data, err := cv.Marshal(obj)
	if err != nil {
		panic("marshal" + err.Error())
	}
	t.Logf("data:%s", data)
}
