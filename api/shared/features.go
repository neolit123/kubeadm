package shared

const (
	// IPv6DualStack is expected to be alpha in v1.16
	IPv6DualStack = "IPv6DualStack"
	// PublicKeysECDSA is expected to be alpha in v1.19
	PublicKeysECDSA = "PublicKeysECDSA"
)

// FeatureEnabled indicates whether a feature name has been enabled
func FeatureEnabled(featureList map[string]bool, featureName string) bool {
	if enabled, ok := featureList[string(featureName)]; ok {
		return enabled
	}
	return false
}
