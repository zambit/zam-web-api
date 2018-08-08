package isc

// IEventNotificator used to notify other services about events which occurs in web API related to an user
type IEventNotificator interface {
	// RegistrationVerificationRequested emitted when user phone registration is required during registration process
	RegistrationVerificationRequested(userID, userPhone, verificationCode string) error

	// RegistrationCompleted emitted when user completes registration process
	RegistrationCompleted(userID, userPhone string) error

	// RegistrationVerificationRequested emitted when user should verify password recovery
	PasswordRecoveryVerificationRequested(userID, userPhone, verificationCode string) error

	// RegistrationCompleted emitted when user completes password recovery
	PasswordRecoveryCompleted(userID, userPhone string) error
}
