package shared

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// TypeMeta ...
type TypeMeta struct {
	// APIVersion ...
	APIVersion string `json:"apiVersion,omitempty"`
	// Kind ...
	Kind string `json:"kind,omitempty"`
}

// Kind ...
type Kind interface {
	Version() string
	Name() string
	ConvertUp(*Converter, Kind) (Kind, error)
	ConvertDown(*Converter, Kind) (Kind, error)
	ConvertUpName() string
	ConvertDownName() string
	Validate(Kind) error
	Default(Kind)
}

// VersionKinds ...
type VersionKinds struct {
	Version string
	Kinds   []Kind
}

// Converter ...
type Converter struct {
	versionKinds []VersionKinds
	cache        map[string]interface{}
}

// NewConverter ...
func NewConverter(versionKinds []VersionKinds) *Converter {
	return &Converter{
		versionKinds: versionKinds,
		cache:        map[string]interface{}{},
	}
}

// GetObjectFromJSON ...
func GetObjectFromJSON(cv *Converter, input []byte) (interface{}, error) {
	typemeta := TypeMeta{}
	if err := json.Unmarshal(input, &typemeta); err != nil {
		return nil, err
	}

	fmt.Println(typemeta.APIVersion, typemeta.Kind)

	gv := strings.Split(typemeta.APIVersion, "/")
	if len(gv) != 2 {
		return nil, fmt.Errorf("malformed group/version: %s", typemeta.APIVersion)
	}

	iface := GetObject(cv, gv[1], typemeta.Kind)
	if iface == nil {
		return nil, fmt.Errorf("no object for version/kind: %s/%s", gv[1], typemeta.Kind)
	}

	if err := json.Unmarshal(input, iface); err != nil {
		return nil, err
	}

	return iface, nil
}

// GetObject ...
func GetObject(cv *Converter, version, kind string) interface{} {
	for _, vk := range cv.versionKinds {
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

// DeepCopy ...
func DeepCopy(src interface{}) interface{} {
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

// ConvertToLatest ...
func ConvertToLatest(cv *Converter, in Kind) (Kind, error) {
	if len(cv.versionKinds) == 0 {
		return nil, fmt.Errorf("no versions to convert to")
	}
	latest := cv.versionKinds[len(cv.versionKinds)-1]
	return ConvertTo(cv, in, latest.Version)
}

// ConvertTo ...
func ConvertTo(cv *Converter, in Kind, targetVersion string) (Kind, error) {
	if len(cv.versionKinds) == 0 {
		return nil, fmt.Errorf("no versions to convert to")
	}

	version := in.Version()
	kind := in.Name()

	fmt.Println("kind", kind, "version", version)

	targetVersionIdx := -1
	for i, vk := range cv.versionKinds {
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
	for i := 0; i < len(cv.versionKinds); i++ {
		vk := cv.versionKinds[i]
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
		vk := cv.versionKinds[i]

		for _, k := range vk.Kinds {
			if k.ConvertDownName() == kind {
				out, err = k.ConvertDown(cv, in)
				if err != nil {
					return nil, fmt.Errorf("cannot convert %s/%s to %s/%s: %v", in.Version(), in.Name(), vk.Version, k.Name(), err)
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
		vk := cv.versionKinds[i]

		// find the same kind in the next version
		for _, k := range vk.Kinds {
			if k.ConvertUpName() == kind {
				in = in.(Kind)
				out, err = k.ConvertUp(cv, in)
				// fmt.Printf("convert up %T, %p - %T, %p\n", in, in, out, out)
				if err != nil {
					return nil, fmt.Errorf("cannot convert %s/%s to %s/%s: %v", in.Version(), in.Name(), vk.Version, k.Name(), err)
				}
				in = out
				kind = k.Name()
			}
		}
	}
	return out, nil
}
