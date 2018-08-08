package isc

import (
	"git.zam.io/wallet-backend/web-api/pkg/services/broker"
)

const (
	resource = "users"

	actionRegistrationVerificationRequired = "registration_verification_required_event"
	actionRegistrationCompleted            = "registration_verification_completed_event"

	actionPasswordRecoveryVerificationRequired = "password_recovery_verification_required_event"
	actionPasswordRecoveryCompleted            = "password_recovery_completed_event"
)

// notificator implements IEventNotificator sending events thought broker according to docs
type notificator struct {
	b broker.IBroker
}

// New notificator instance from given broker
func New(b broker.IBroker) IEventNotificator {
	return &notificator{b}
}

// RegistrationVerificationRequested
func (n notificator) RegistrationVerificationRequested(userID, userPhone, verificationCode string) error {
	return n.b.Publish(identifier(actionRegistrationVerificationRequired, userID), pl{
		"user_id":           userID,
		"user_phone":        userPhone,
		"verification_code": verificationCode,
	})
}

// RegistrationCompleted
func (n notificator) RegistrationCompleted(userID, userPhone string) error {
	return n.b.Publish(identifier(actionRegistrationCompleted, userID), pl{
		"user_id":    userID,
		"user_phone": userPhone,
	})
}

// PasswordRecoveryVerificationRequested
func (n notificator) PasswordRecoveryVerificationRequested(userID, userPhone, verificationCode string) error {
	return n.b.Publish(identifier(actionPasswordRecoveryVerificationRequired, userID), pl{
		"user_id":       userID,
		"user_phone":    userPhone,
		"recovery_code": verificationCode,
	})
}

// PasswordRecoveryCompleted
func (n notificator) PasswordRecoveryCompleted(userID, userPhone string) error {
	return n.b.Publish(identifier(actionPasswordRecoveryCompleted, userID), pl{
		"user_id":    userID,
		"user_phone": userPhone,
	})
}

func identifier(action, id string) broker.Identifier {
	return broker.Identifier{
		Resource: resource,
		Action:   action,
		ID:       id,
	}
}

type pl map[string]interface{}
