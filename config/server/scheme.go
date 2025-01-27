package server

import (
	"time"
)

// AuthScheme web-authorization related parameters
type AuthScheme struct {
	// TokenName specifies token prefix in Authorization header
	TokenName string

	// TokenExpire authorization token live duration before become expire (example: 24h45m15s)
	TokenExpire time.Duration

	// TokenType describes token storage type.
	//
	// Possible values:
	//
	//  mem - inmemory token storage
	//
	//  jwt - jwt token storage
	//
	//  jwtpersisten - jwt token storage which uses persistent storage for token validation
	TokenStorage string

	SignUpTokenExpire time.Duration
	SignUpRetryDelay  time.Duration
}

// StorageScheme holds values specific for nosql storage
type StorageScheme struct {
	// URI used to connect to the storage.
	//
	// Possible schemes:
	//
	//  mem:// - in-memory storage
	//
	//  redis:// or rediss:// - redis storage, also supports redis cluster passing hosts slitted by comma
	URI string
}

// GeneratorScheme
type GeneratorScheme struct {
	// CodeLen desired length of generated code
	CodeLen int

	// CodeAlphabet sets of letters used to generate code
	CodeAlphabet string
}

// NotificatorURL specifies notificator URI which is used to determine actual implementation.
type NotificatorScheme struct {
	// Possible schemes (empty value will cause to simply log records using log configuration):
	//
	// 	 https://{twilio_sid}:{twilio_token}@api.twilio.com/?From={send_from_phone} - using twilio sms service
	//
	//   https://hooks.slack.com/services/{hook_part} - using slack "https://hooks.slack.com/services/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"
	//
	//   file://{path_to_file} - using file to log all notifications records. File will be created, if not exists.
	// Make sure you have enough rights!
	URL string
}

// Scheme web-server params
type Scheme struct {
	// Host to listen on such address, accept both ip4 and ip6 addresses
	Host string

	// Port to listen on, negative values will cause UB
	Port int

	// JWT specific configuration, there is no default values, so if token jwt like storage is used, this must be defined
	JWT *struct {
		Secret string
		Method string
	}

	// Auth
	Auth AuthScheme

	// Storage
	Storage StorageScheme

	// Generator
	Generator GeneratorScheme

	// Notificator
	Notificator NotificatorScheme
}
