package server

import (
	"time"
)

// AuthScheme web-authorization related parameters
type AuthScheme struct {
	// TokenName specifies token prefix in Authorization header
	TokenName   string

	// TokenExpire authorization token live duration before become expire (example: 24h45m15s)
	TokenExpire time.Duration

	SignUpTokenExpire time.Duration
	SignUpRetryDelay  time.Duration
}

// StorageScheme holds values specific for nosql storage
type StorageScheme struct {
	// URI used to connect to the storage, now only redis is supported, leave empty to use in-mem storage
	URI string
}

// Scheme web-server params
type Scheme struct {
	// Host to listen on such address, accept both ip4 and ip6 addresses
	Host string

	// Port to listen on, negative values will cause UB
	Port int

	// Auth
	Auth AuthScheme

	// Storage
	Storage StorageScheme

	// GeneratorType specifies generator type (now only "mem" and "" allowed)
	GeneratorType string
}
