package providers

import (
	"git.zam.io/wallet-backend/common/pkg/types"
	"git.zam.io/wallet-backend/web-api/config/logging"
	"git.zam.io/wallet-backend/web-api/pkg/services/sentry"
	"git.zam.io/wallet-backend/web-api/pkg/services/sentry/raven"
	"github.com/sirupsen/logrus"
	"reflect"
)

// Reporter
func Reporter(logger logrus.FieldLogger, env types.Environment, cfg logging.Scheme) (sentry.IReporter, error) {
	if len(cfg.ErrorReporter.DSN) == 0 {
		logger.Info("using default sentry reporter due to config value is empty")
		return sentry.Global(), nil
	} else {
		r, err := raven.New(cfg.ErrorReporter.DSN, string(env))
		if err != nil {
			logger.WithError(err).Error("failed to initialize sentry error reporter")
			return nil, err
		}

		logger.Infof("sentry logger reporter has been created with type %s", reflect.TypeOf(r).Name())

		return r, nil
	}
}
