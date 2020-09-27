package v1beta3

import "errors"

// ValidateFoo ...
func ValidateFoo(in *Foo) error {
	if len(in.A) == 0 {
		return errors.New("A must be set")
	}
	if len(in.B) == 0 {
		return errors.New("B must be set")
	}
	if in.N != nil {
		if len(in.N.P) == 0 {
			return errors.New("P must be set")
		}
	}
	return nil
}
