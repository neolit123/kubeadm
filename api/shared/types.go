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
	Validate() error
	Default() error
	GetTypeMeta() *TypeMeta
}

// VersionKinds ...
type VersionKinds struct {
	Version string
	Kinds   []Kind
}

// Converter ...
type Converter struct {
	group         string
	versionKinds  []VersionKinds
	cache         map[string]Kind
	unmarshalFunc func([]byte, interface{}) error
	marshalFunc   func(interface{}) ([]byte, error)
}

// NewConverter ...
func NewConverter(group string, versionKinds []VersionKinds) *Converter {
	return &Converter{
		group:        group,
		versionKinds: versionKinds,
		cache:        map[string]Kind{},
	}
}

// GetGroup ...
func (cv *Converter) GetGroup() string {
	return cv.group
}

// AddToCache ...
func (cv *Converter) AddToCache(key string, kind Kind) {
	cv.cache[key] = DeepCopy(kind)
}

// GetFromCache ...
func (cv *Converter) GetFromCache(key string) Kind {
	return DeepCopy(cv.cache[key])
}

// GetObjectFromBytes ...
func (cv *Converter) GetObjectFromBytes(typemeta *TypeMeta, input []byte) (Kind, error) {
	kind, err := cv.GetObject(typemeta)
	if err != nil {
		return nil, err
	}

	if err := cv.unmarshalFunc(input, kind); err != nil {
		return nil, err
	}

	return kind, nil
}

// GetObject ...
func (cv *Converter) GetObject(typemeta *TypeMeta) (Kind, error) {
	gv := strings.Split(typemeta.APIVersion, "/")
	if len(gv) != 2 {
		return nil, fmt.Errorf("malformed group/version: %s", typemeta.APIVersion)
	}

	for _, vk := range cv.versionKinds {
		if gv[1] != vk.Version {
			continue
		}
		for _, k := range vk.Kinds {
			if typemeta.Kind != k.Name() {
				continue
			}
			t := reflect.TypeOf(k)
			kind := (reflect.New(t.Elem()).Interface()).(Kind)
			cv.SetTypeMeta(kind)
			return kind, nil
		}
	}
	return nil, fmt.Errorf("no object for: %+v", typemeta)
}

// DeepCopy ...
func DeepCopy(src Kind) Kind {
	if src == nil {
		panic("nil value passed to DeepCopy")
	}
	bytes, err := json.Marshal(src)
	if err != nil {
		panic("error marshal")
	}
	t := reflect.TypeOf(src)
	dst := (reflect.New(t.Elem()).Interface()).(Kind)
	if err := json.Unmarshal(bytes, dst); err != nil {
		panic("error unmarshal: " + err.Error())
	}
	return dst
}

// ConvertToLatest ...
func (cv *Converter) ConvertToLatest(in Kind) (Kind, error) {
	if len(cv.versionKinds) == 0 {
		return nil, fmt.Errorf("no versions to convert to")
	}
	latest := cv.versionKinds[len(cv.versionKinds)-1]
	return cv.ConvertTo(in, latest.Version)
}

// ConvertTo ...
func (cv *Converter) ConvertTo(in Kind, targetVersion string) (Kind, error) {
	if len(cv.versionKinds) == 0 {
		return nil, fmt.Errorf("no versions to convert to")
	}

	version := in.Version()
	kind := in.Name()

	targetVersionIdx := -1
	for i, vk := range cv.versionKinds {
		if targetVersion == vk.Version {
			targetVersionIdx = i
			break
		}
	}

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

	// already target version
	if versionIdx == targetVersionIdx {
		return in, nil
	}

	var out = in
	var err error

	if versionIdx < targetVersionIdx {
		// fmt.Println("convert up")
		goto convertUp
	}

	// fmt.Println("convert down")
	for i := versionIdx; i >= targetVersionIdx; i-- {
		vk := cv.versionKinds[i]

		for _, k := range vk.Kinds {
			if k.ConvertDownName() == kind {
				out, err = k.ConvertDown(cv, in)
				if err != nil {
					return nil, fmt.Errorf("cannot convert %s/%s to %s/%s: %v", in.Version(), in.Name(), vk.Version, k.Name(), err)
				}
				in = out
				kind = k.ConvertUpName()
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
				out, err = k.ConvertUp(cv, in)
				if err != nil {
					return nil, fmt.Errorf("cannot convert %s/%s to %s/%s: %v", in.Version(), in.Name(), vk.Version, k.Name(), err)
				}
				in = out
				kind = k.ConvertDownName()
			}
		}
	}
	return out, nil
}

// GetTypeMetaFromBytes ...
func (cv *Converter) GetTypeMetaFromBytes(input []byte) (*TypeMeta, error) {
	if cv.unmarshalFunc == nil {
		return nil, fmt.Errorf("unmarshal function not set")
	}

	typemeta := &TypeMeta{}
	if err := cv.unmarshalFunc(input, typemeta); err != nil {
		return nil, fmt.Errorf("cannot get TypeMeta: %v", err)
	}
	return typemeta, nil
}

// SetTypeMeta ...
func (cv *Converter) SetTypeMeta(kind Kind) {
	typemeta := kind.GetTypeMeta()
	typemeta.APIVersion = cv.group + "/" + kind.Version()
	typemeta.Kind = kind.ConvertDownName()
}

// SetMarshalFunc ...
func (cv *Converter) SetMarshalFunc(f func(interface{}) ([]byte, error)) {
	cv.marshalFunc = f
}

// SetUnmarshalFunc ...
func (cv *Converter) SetUnmarshalFunc(f func([]byte, interface{}) error) {
	cv.unmarshalFunc = f
}

// Marshal ...
func (cv *Converter) Marshal(k Kind) ([]byte, error) {
	if cv.marshalFunc == nil {
		return nil, fmt.Errorf("marshal function not set")
	}
	return cv.marshalFunc(k)
}

// Unmarshal ...
func (cv *Converter) Unmarshal(b []byte, k Kind) error {
	if cv.unmarshalFunc == nil {
		return fmt.Errorf("unmarshal function not set")
	}
	return cv.unmarshalFunc(b, k)
}