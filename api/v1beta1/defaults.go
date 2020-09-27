package v1beta1

const (
	// DefaultA ...
	DefaultA = "defaultA"
	// DefaultB ...
	DefaultB = "defaultB"
)

// DefaultFoo ...
func DefaultFoo(in *Foo) *Foo {
	if len(in.A) == 0 {
		in.A = DefaultA
	}
	if len(in.B) == 0 {
		in.B = DefaultB
	}
	return in
}
