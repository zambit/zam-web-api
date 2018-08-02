package stub

import (
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications"
	"github.com/sirupsen/logrus"
)

//
type stubNotificator struct {
	logger logrus.FieldLogger
}

//
func New(logger logrus.FieldLogger) notifications.ISender {
	return stubNotificator{
		logger: logger.WithField("module", "stub_notificator"),
	}
}

func (n stubNotificator) Send(action string, data interface{}, level notifications.ImportanceLevel) error {
	n.logger.WithFields(
		logrus.Fields{
			"action": action,
			"level":  level,
			"data":   data,
		},
	).Infof("notification send")
	return nil
}
