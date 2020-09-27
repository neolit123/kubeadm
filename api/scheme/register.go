package scheme

const (
	schemeGroup = "kubeadm.k8s.io"
	// V1Beta1 ...
	V1Beta1 = "v1beta1"
	// V1Beta1Foo ...
	V1Beta1Foo = "Foo"

	// V1Beta2 ...
	V1Beta2 = "v1beta2"
	// V1Beta2Foo ...
	V1Beta2Foo = "Foo"
	// V1Beta2Bar ...
	V1Beta2Bar = "Bar"

	// V1Beta3 ...
	V1Beta3 = "v1beta3"
	// V1Beta3Foo ...
	V1Beta3Foo = "Foo"
	// V1Beta3Bar ...
	V1Beta3Bar = "Bar"
)

var versionKinds = map[string][]string{
	V1Beta1: []string{
		"Foo",
	},
	V1Beta2: []string{
		"Foo",
		"Bar",
	},
	V1Beta3: []string{
		"Foo",
		"Bar",
	},
}
