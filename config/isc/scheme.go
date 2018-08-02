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
}
