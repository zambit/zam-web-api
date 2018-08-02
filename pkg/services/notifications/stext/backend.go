package stext

// IBackend used to sent messages using simple-text notifier
type IBackend interface {
	Send(recipient string, body string) error
}
