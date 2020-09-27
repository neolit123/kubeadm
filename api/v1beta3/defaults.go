package v1beta3

const (
	// DefaultA ...
	DefaultA = "defaultA"
	// DefaultB ...
	DefaultB = "defaultB"
	// DefaultC ...
	DefaultC = "defaultC"
	// DefaultBarP ...
	DefaultBarP = "defaultBarP"
)

// DefaultFoo ...
func DefaultFoo(in *Foo) *Foo {
	if len(in.A) == 0 {
		in.A = DefaultA
	}
	if len(in.B) == 0 {
		in.B = DefaultB
	}
	if in.N == nil {
		in.N = &Bar{
			P: DefaultBarP,
		}
	}
	return in
}
