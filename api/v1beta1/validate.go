package v1beta1

import "errors"

// ValidateFoo ...
func ValidateFoo(in *Foo) error {
	if len(in.A) == 0 {
		return errors.New("A must be set")
	}
	if len(in.B) == 0 {
		return errors.New("B must be set")
	}
	return nil
}
