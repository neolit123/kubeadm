package v1beta2

const (
	// DefaultA ...
	DefaultA = "defaultA"
	// DefaultB ...
	DefaultB = "defaultB"
	// DefaultC ...
	DefaultC = "defaultC"
	// DefaultBarP ...
	DefaultBarP = "defaultBarP"
	// DefaultBarQ ...
	DefaultBarQ = "defaultBarQ"
)

// DefaultFoo ...
func DefaultFoo(in *Foo) *Foo {
	if len(in.A) == 0 {
		in.A = DefaultA
	}
	if len(in.B) == 0 {
		in.B = DefaultB
	}
	if len(in.C) == 0 {
		in.C = DefaultC
	}
	if in.N == nil {
		in.N = &Bar{
			P: DefaultBarP,
			Q: DefaultBarQ,
		}
	}
	return in
}
