package raven

import (
	"git.zam.io/wallet-backend/web-api/pkg/services/sentry"
	"github.com/getsentry/raven-go"
)

// Reporter implements IReporter using raven sentry client
type Reporter struct {
	client *raven.Client
}

// New
func New(dsn, environment string) (sentry.IReporter, error) {
	client, err := raven.New(dsn)
	if err != nil {
		return nil, err
	}

	r := &Reporter{
		client: client,
	}
	r.client.SetEnvironment(environment)

	return r, nil
}

// ReportErr
func (r *Reporter) ReportErr(err error, tags map[string]string) {
	r.client.CaptureError(err, tags)
}

// CapturePanic
func (r *Reporter) CapturePanic(cb func(), tags map[string]string) {
	r.client.CapturePanic(cb, tags)
}
