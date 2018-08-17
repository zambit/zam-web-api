package isc

// Scheme defines inter-services communication primarily using MQ broker
type Scheme struct {
	// BrokerURI mq broker url
	//
	// Supports only redis for now, expects ordinal redis connection url
	BrokerURI string

	// ServeStats
	StatsEnabled bool

	// StatsPath
	StatsPath string

	// WalletApiDiscovery holds wallet-api discovery settings
	WalletApiDiscovery DiscoveryScheme
}

// DiscoveryScheme holds settings which describes access to internal service api's
type DiscoveryScheme struct {
	// Host
	Host string

	// AccessToken which wallet-api requires
	AccessToken string
}
