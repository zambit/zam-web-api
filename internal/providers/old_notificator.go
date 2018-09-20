package providers

import (
	serverconf "git.zam.io/wallet-backend/web-api/config/server"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications/stext/factory"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications/stub"
	"github.com/sirupsen/logrus"
)

// Notificator
func Notificator(conf serverconf.Scheme, logger logrus.FieldLogger) (notifications.ISender, error) {
	if conf.Notificator.URL == "" {
		return stub.New(logger), nil
	} else {
		return factory.New(conf.Notificator.URL)
	}
}
