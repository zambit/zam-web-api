package providers

import (
	"git.zam.io/wallet-backend/web-api/pkg/services/broker"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications"
	"git.zam.io/wallet-backend/web-api/internal/services/isc"
	"github.com/sirupsen/logrus"
)

// EventNotificator
func EventNotificator(
	broker broker.IBroker,
	oldNotificator notifications.ISender,
	logger logrus.FieldLogger,
) (brokerNotificator isc.IEventNotificator) {
	if broker != nil {
		brokerNotificator = isc.New(broker)
	} else {
		brokerNotificator = isc.NewStub(logger)
	}
	// provide merged notificator due to obscene of real notifications service
	return isc.NewMerged(brokerNotificator, oldNotificator)
}
