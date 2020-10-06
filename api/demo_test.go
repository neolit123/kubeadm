package main

import (
	"encoding/json"
	"testing"

	"k8s.io/kubeadm/api/scheme"
	"k8s.io/kubeadm/api/shared"
)

var input = []byte(`
{
	"apiVersion": "kubeadm.k8s.io/v1beta1",
	"kind": "Foo",
	"a": "aaa",
	"b": "bbb",
	"c": "ccc"
}
`)

func TestDemo(t *testing.T) {
	cv := shared.NewConverter(scheme.Group, scheme.VersionKinds)
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

	obj, _ = cv.ConvertTo(obj, "v1beta3")
	t.Logf("--------%#v", obj)

	obj, _ = cv.ConvertTo(obj, "v1beta1")
	t.Logf("--------%#v", obj)

	obj, _ = cv.ConvertTo(obj, "v1beta2")
	t.Logf("--------%#v", obj)

	obj, err = obj.ConvertDown(cv, obj)
	t.Logf("--------%#v", obj)

	data, err := cv.Marshal(obj)
	if err != nil {
		panic("marshal" + err.Error())
	}
	t.Logf("data:%s", data)
}
