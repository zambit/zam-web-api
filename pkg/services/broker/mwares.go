package broker

import "git.zam.io/wallet-backend/web-api/pkg/services/sentry"

// NewReportMiddleware captures panics and returned errors and logs them with reporter
func NewReportMiddleware(reporter sentry.IReporter, tags map[string]string) MiddlewareFunc {
	return func(b IBroker, d Delivery, next ConsumeFunc) (err error) {
		reporter.CapturePanic(func() {
			err = next(b, d)
			if err != nil {
				reporter.ReportErr(err, tags)
			}
		}, tags)
		return err
	}
}
