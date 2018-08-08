package isc

import (
	"git.zam.io/wallet-backend/web-api/internal/services/notifications"
)

// mergedNotificator wraps event notificator and old notificator sending notification simultaneously
type mergedNotificator struct {
	eventNotificator IEventNotificator
	oldNotificator   notifications.ISender
}

// NewMerged merged notificator
func NewMerged(eventNotificator IEventNotificator, oldNotificator notifications.ISender) IEventNotificator {
	return &mergedNotificator{eventNotificator, oldNotificator}
}

func (n *mergedNotificator) RegistrationVerificationRequested(userID, userPhone, verificationCode string) error {
	err := n.oldNotificator.Send(
		notifications.ActionRegistrationConfirmationRequested,
		map[string]interface{}{
			"phone": userPhone,
			"code":  verificationCode,
		},
		notifications.Urgent,
	)
	if err != nil {
		return err
	}
	return n.eventNotificator.RegistrationVerificationRequested(userID, userPhone, verificationCode)
}

func (n *mergedNotificator) RegistrationCompleted(userID, userPhone string) error {
	err := n.oldNotificator.Send(
		notifications.ActionRegistrationCompleted,
		map[string]interface{}{
			"id": userID,
		},
		notifications.Urgent,
	)
	if err != nil {
		return err
	}
	return n.eventNotificator.RegistrationCompleted(userID, userPhone)
}

func (n *mergedNotificator) PasswordRecoveryVerificationRequested(userID, userPhone, verificationCode string) error {
	err := n.oldNotificator.Send(
		notifications.ActionPasswordRecoveryConfirmationRequested,
		map[string]interface{}{
			"phone": userPhone,
			"code":  verificationCode,
		},
		notifications.Urgent,
	)
	if err != nil {
		return err
	}
	return n.eventNotificator.PasswordRecoveryVerificationRequested(userID, userPhone, verificationCode)
}

func (n *mergedNotificator) PasswordRecoveryCompleted(userID, userPhone string) error {
	err := n.oldNotificator.Send(
		notifications.ActionPasswordRecoveryCompleted,
		map[string]interface{}{
			"id": userID,
		},
		notifications.Urgent,
	)
	if err != nil {
		return err
	}
	return n.eventNotificator.PasswordRecoveryCompleted(userID, userPhone)
}
