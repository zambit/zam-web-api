package logging

// Scheme describes application logging
type Scheme struct {
	// LogLevel specifies log level, one of:
	// - debug
	// - info
	// - warning
	// - error
	// - fatal
	// - panic
	//
	// Wrong values will be ignored and default one applied.
	//
	// Default is: info
	LogLevel string

	// ErrorReporter configure
	ErrorReporter struct {
		// DSN is sentry.io DSN URL
		DSN string
	}
}
