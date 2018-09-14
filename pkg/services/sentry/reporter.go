package sentry

// IReporter
type IReporter interface {
	ReportErr(err error, tags map[string]string)

	CapturePanic(cb func(), tags map[string]string)
}

var singleton IReporter = noopReporter{}

// Global get global instance
func Global() IReporter {
	return singleton
}

// SetGlobal use with case: it isn't thread-safe
func SetGlobal(r IReporter) {
	singleton = r
}

// noopReporter no-op IReporter implementation
type noopReporter struct{}

// ReportErr implements IReporter
func (noopReporter) ReportErr(err error, tags map[string]string) {}

// CapturePanic implements IReporter
func (noopReporter) CapturePanic(cb func(), tags map[string]string) { cb() }
