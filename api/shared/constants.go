package shared

const (
	// MinimumAddressesInServiceSubnet defines minimum amount of nodes the Service subnet should allow.
	// We need at least ten, because the DNS service is always at the tenth cluster clusterIP
	MinimumAddressesInServiceSubnet = 10

	// TokenStr flags sets both the discovery-token and the tls-bootstrap-token when those values are not provided
	TokenStr = "token"

	// TokenTTL flag sets the time to live for token
	TokenTTL = "token-ttl"

	// TokenUsages flag sets the usages of the token
	TokenUsages = "usages"

	// TokenGroups flag sets the authentication groups of the token
	TokenGroups = "groups"
)
