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
}

// Scheme web-server params
type Scheme struct {
	// Host to listen on such address, accept both ip4 and ip6 addresses
	Host string

	// Port to listen on, negative values will cause UB
	Port int

	// Auth
	Auth AuthScheme
}
