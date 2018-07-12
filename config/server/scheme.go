package server

import (
	"time"
)

// AuthScheme describes global authorization params
type AuthScheme struct {
	TokenName   string
	TokenExpire time.Duration
}

// Scheme describes server listen params
type Scheme struct {
	Host string
	Port int

	Auth AuthScheme
}
