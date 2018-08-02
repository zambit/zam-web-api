package isc

import (
	"github.com/sirupsen/logrus"
)

// stubNotificator logs all events using logger
type stubNotificator struct {
	logger logrus.FieldLogger
}

// NewStub
func NewStub(logger logrus.FieldLogger) IEventNotificator {
	return &stubNotificator{logger.WithField("module", "notificator.stub")}
}

func (n stubNotificator) RegistrationVerificationRequested(userID, userPhone, verificationCode string) error {
	n.logger.WithFields(logrus.Fields{
		"user_id":           userID,
		"user_phone":        userPhone,
		"verification_code": verificationCode,
	}).Info("user registration verification required")
	return nil
}

func (n stubNotificator) RegistrationCompleted(userID string) error {
	n.logger.WithField("user_id", userID).Info("user registration completed")
	return nil
}

func (n stubNotificator) PasswordRecoveryVerificationRequested(userID, userPhone, verificationCode string) error {
	n.logger.WithFields(logrus.Fields{
		"user_id":           userID,
		"user_phone":        userPhone,
		"recovery_code": verificationCode,
	}).Info("user password recovery verification required")
	return nil
}

func (n stubNotificator) PasswordRecoveryCompleted(userID string) error {
	n.logger.WithField("user_id", userID).Info("user password recovery completed")
	return nil
}
