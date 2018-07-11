package server

import (
	"time"
)

// Scheme describes server listen params
type Scheme struct {
	Host string
	Port int

	Auth struct {
		TokenName   string
		TokenExpire time.Duration
	}
}
