package notifications

// ITransport used to delivery user notifications in different ways
type ITransport interface {
	Send(recipient string, body string) error
}
