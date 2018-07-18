package notifications

// ImportanceLevel represent how much given notification is important
type ImportanceLevel int

func (level ImportanceLevel) String() string {
	switch level {
	case Urgent:
		return "urgent"
	case Ordinal:
		return "ordinal"
	default:
		panic("wrong importance level")
	}
}

const (
	Urgent ImportanceLevel = iota
	Confirmation
	Ordinal
)

const (
	// ActionRegistrationConfirmationRequested requires service to send confirmation
	// This actions requires "phone" and "code" to be specified in data map
	ActionRegistrationConfirmationRequested = "action_registration_confirmation_requested"

	// ActionRegistrationCompleted notify user about successful registration and request confirmation
	ActionRegistrationCompleted = "action_registration_completed"
)

// ISender intends to perform all notification actions depending on user settings
type ISender interface {
	// Send notification representing given action with some required data and importance level
	Send(action string, data interface{}, level ImportanceLevel) error
}
