package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	// "sigs.k8s.io/yaml"
)

type Kind interface {
	Version() string
	Name() string
	ConvertUp(*manager, interface{}) (interface{}, error)
	ConvertDown(*manager, interface{}) (interface{}, error)
	ConvertUpName() string
	ConvertDownName() string
	Validate() error
	Default()
}

type VersionKinds struct {
	Version string
	Kinds   []Kind
}

// -------------------------

type Foo struct {
	A string
	B string
}

func (*Foo) ConvertUp(man *manager, in interface{}) (interface{}, error) {
	return nil, nil
}

func (*Foo) ConvertDown(man *manager, in interface{}) (interface{}, error) {
	// fmt.Printf("ConvertDown %T, %p\n", in, in)
	x, ok := in.(*Bar)
	if !ok {
		return nil, fmt.Errorf("cannot convert input to Bar")
	}
	obj := &Foo{}
	man.cache["Bar"] = deepCopy(in) // CACHE Bar
	// fmt.Printf("storing %T %p %#v\n", in, in, man.cache)
	obj.A = x.A
	obj.B = x.B
	return obj, nil
}

func (*Foo) ConvertUpName() string {
	return "Foo"
}

func (*Foo) ConvertDownName() string {
	return "Bar"
}

func (*Foo) Validate() error {
	return nil
}

func (*Foo) Default() {
}

func (*Foo) Version() string {
	return "v1beta1"
}

func (*Foo) Name() string {
	return "Foo"
}

// -------------

type Bar struct {
	A string
	B string
	C string
}

func (*Bar) ConvertDown(man *manager, in interface{}) (interface{}, error) {
	fmt.Printf("ConvertDown %T, %p\n", in, in)
	x, ok := in.(*Zed)
	if !ok {
		return nil, fmt.Errorf("cannot convert input to *Zed")
	}
	obj := &Bar{}
	man.cache["Zed"] = deepCopy(in) // CACHE Zed
	// fmt.Printf("storing %T %p %#v\n", in, in, man.cache)
	obj.A = x.A
	obj.B = x.B
	obj.C = x.C
	return obj, nil
}

func (bar *Bar) ConvertUp(man *manager, in interface{}) (interface{}, error) {
	// fmt.Printf("ConvertUp %T, %p\n", in, in)
	x, ok := in.(*Foo)
	if !ok {
		return nil, fmt.Errorf("cannot convert input to *Foo")
	}

	cv, ok := man.cache["Bar"]
	if !ok {
		return nil, fmt.Errorf("cannot fetch from cache")
	}
	b, ok := cv.(*Bar)
	if !ok {
		return nil, fmt.Errorf("cannot convert cache value")
	}

	b.A = x.A
	b.B = x.B
	// b.C = b.C RES
	//delete(man.cache, in)
	// fmt.Printf("------- %p\n", b)
	return b, nil
}

func (*Bar) ConvertDownName() string {
	return "Zed"
}

func (*Bar) ConvertUpName() string {
	return "Foo"
}

func (*Bar) Validate() error {
	return nil
}

func (*Bar) Default() {
}

func (*Bar) Version() string {
	return "v1beta2"
}

func (*Bar) Name() string {
	return "Bar"
}

// ---------------

type Zed struct {
	A string
	B string
	C string
	D string
}

func (*Zed) ConvertDown(man *manager, in interface{}) (interface{}, error) {
	// fmt.Printf("ConvertDown %T, %p\n", in, in)
	x, ok := in.(*Zed)
	if !ok {
		return nil, fmt.Errorf("cannot convert input to *Zed")
	}
	obj := &Bar{}
	man.cache["Zed"] = deepCopy(in)
	obj.A = x.A
	obj.B = x.B
	obj.C = x.C
	return obj, nil
}

func (*Zed) ConvertUp(man *manager, in interface{}) (interface{}, error) {
	// fmt.Printf("ConvertUp %T, %p\n", in, in)
	x, ok := in.(*Bar)
	if !ok {
		return nil, fmt.Errorf("cannot convert input to *Bar")
	}

	// fmt.Printf("--- %p, cache %#v\n", in, man.cache)

	cv, ok := man.cache["Zed"]
	if !ok {
		return nil, fmt.Errorf("cannot fetch from cache")
	}
	b, ok := cv.(*Zed)
	if !ok {
		return nil, fmt.Errorf("cannot convert cache value")
	}

	b.A = x.A
	b.B = x.B
	b.C = x.C
	// b.D = b.D // RESTORE FROM CACHE
	// delete(man.cache, in)
	return b, nil
}

func (*Zed) ConvertUpName() string {
	return "Bar"
}

func (*Zed) ConvertDownName() string {
	return "Bar"
}

func (*Zed) Validate() error {
	return nil
}

func (*Zed) Default() {
}

func (*Zed) Version() string {
	return "v1beta3"
}

func (*Zed) Name() string {
	return "Zed"
}

// ------------

var defaultVersionKinds = []VersionKinds{
	{"v1beta1", []Kind{&Foo{}}},
	{"v1beta2", []Kind{&Bar{}}},
	{"v1beta3", []Kind{&Zed{}}},
}

type manager struct {
	versionKinds []VersionKinds
	cache        map[string]interface{}
}

func newManager() *manager {
	v := make([]VersionKinds, len(defaultVersionKinds))
	copy(v, defaultVersionKinds)
	return &manager{
		versionKinds: v,
		cache:        map[string]interface{}{},
	}
}

func getObject(man *manager, version, kind string) interface{} {
	for _, vk := range man.versionKinds {
		if version != vk.Version {
			continue
		}
		for _, k := range vk.Kinds {
			if kind != k.Name() {
				continue
			}
			t := reflect.TypeOf(k)
			return reflect.New(t.Elem()).Interface()
		}
	}
	return nil
}

func deepCopy(src interface{}) interface{} {
	if src == nil {
		panic("nil value passed to deepCopy")
	}
	bytes, err := json.Marshal(src)
	if err != nil {
		panic("error marshal")
	}
	t := reflect.TypeOf(src)
	dst := reflect.New(t.Elem()).Interface()
	err = json.Unmarshal(bytes, dst)
	if err != nil {
		panic("error unmarshal: " + err.Error())
	}
	return dst
}

func convertToLatest(man *manager, in interface{}) (interface{}, error) {
	if len(man.versionKinds) == 0 {
		return nil, fmt.Errorf("no versions to convert to")
	}
	latest := man.versionKinds[len(man.versionKinds)-1]
	return convertTo(man, in, latest.Version)
}

func convertTo(man *manager, in interface{}, targetVersion string) (interface{}, error) {
	inKind, ok := in.(Kind)
	if !ok {
		return nil, fmt.Errorf("not a Kind")
	}
	if len(man.versionKinds) == 0 {
		return nil, fmt.Errorf("no versions to convert to")
	}

	version := inKind.Version()
	kind := inKind.Name()

	fmt.Println("kind", kind, "version", version)

	targetVersionIdx := -1
	for i, vk := range man.versionKinds {
		if targetVersion == vk.Version {
			targetVersionIdx = i
			break
		}
	}

	// fmt.Println("targetVersionIdx", targetVersionIdx)

	if targetVersionIdx == -1 {
		return nil, fmt.Errorf("unknown target version %s", targetVersion)
	}

	versionIdx := -1
	for i := 0; i < len(man.versionKinds); i++ {
		vk := man.versionKinds[i]
		if version == vk.Version {
			versionIdx = i
			break
		}
	}
	if versionIdx == -1 {
		return nil, fmt.Errorf("unknown version %s", version)
	}

	fmt.Println("vidx", versionIdx, "targetIdx", targetVersionIdx)

	// already target version
	if versionIdx == targetVersionIdx {
		return in, nil
	}

	var out = in
	var err error

	if versionIdx < targetVersionIdx {
		fmt.Println("convert up")
		goto convertUp
	}

	fmt.Println("convert down")
	for i := versionIdx - 1; i >= targetVersionIdx; i-- {
		vk := man.versionKinds[i]

		for _, k := range vk.Kinds {
			if k.ConvertDownName() == kind {
				inKind = in.(Kind)
				out, err = k.ConvertDown(man, in)
				if err != nil {
					return nil, fmt.Errorf("cannot convert %s/%s to %s/%s: %v", inKind.Version(), inKind.Name(), vk.Version, k.Name(), err)
				}
				in = out
				// fmt.Printf("convert down %T, %p\n", in, in)
				kind = k.Name()
			}
		}
	}
	return out, nil
convertUp:
	for i := versionIdx + 1; i < targetVersionIdx+1; i++ {
		vk := man.versionKinds[i]

		// find the same kind in the next version
		for _, k := range vk.Kinds {
			if k.ConvertUpName() == kind {
				inKind = in.(Kind)
				out, err = k.ConvertUp(man, in)
				// fmt.Printf("convert up %T, %p - %T, %p\n", in, in, out, out)
				if err != nil {
					return nil, fmt.Errorf("cannot convert %s/%s to %s/%s: %v", inKind.Version(), inKind.Name(), vk.Version, k.Name(), err)
				}
				in = out
				kind = k.Name()
			}
		}
	}
	return out, nil
}

func main() {
	man := newManager()
	gvk := struct {
		APIVersion string
		Kind       string
	}{}
	b := []byte(`{"kind": "Zed", "apiVersion": "kubeadm.com/v1beta3", "a": "aa", "b": "bb", "c": "cc", "d": "dd"}`)
	err := json.Unmarshal(b, &gvk)
	if err != nil {
		panic(err)
	}
	iface := getObject(man, strings.Split(gvk.APIVersion, "/")[1], gvk.Kind)
	if iface == nil {
		panic("nil")
	}
	if err := json.Unmarshal(b, iface); err != nil {
		panic(err)
	}

	fmt.Printf("0 %#v\n", iface)

	n, err := convertTo(man, iface, "v1beta2")
	if err != nil {
		fmt.Printf("convert error %v\n", err)
	}

	fmt.Printf("1 %#v\n", n)
	(n.(*Bar)).C = "222"

	z, err := convertTo(man, n, "v1beta3")
	if err != nil {
		fmt.Printf("convert error %v\n", err)
	}
	fmt.Printf("2 %#v\n", z)
}
