package notifications

// IGenerator holds different auxiliary functions for generating purposes
type IGenerator interface {
	// RandomCode generates random confirmation-like code
	RandomCode() string

	// RandomToken generates random token-like char sequence
	RandomToken() string
}
