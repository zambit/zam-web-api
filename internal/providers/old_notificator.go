package providers

import (
	serverconf "git.zam.io/wallet-backend/web-api/config/server"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications/stext/factory"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications/stub"
	"github.com/sirupsen/logrus"
)

// Notificator
func Notificator(conf serverconf.Scheme, logger logrus.FieldLogger) notifications.ISender {
	if conf.NotificatorURL == "" {
		return stub.New(logger)
	} else {
		return factory.New(conf.NotificatorURL)
	}
}
