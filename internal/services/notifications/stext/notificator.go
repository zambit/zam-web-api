package stext

import (
	"fmt"
	old_notifications "git.zam.io/wallet-backend/web-api/internal/services/notifications"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications"
	"github.com/chonla/format"
	"github.com/pkg/errors"
)

// sender renders notification in simple-text form and send it to a recipient
type sender struct {
	backend notifications.ITransport
}

func New(backend notifications.ITransport) old_notifications.ISender {
	return &sender{backend: backend}
}

// Send renders and sends notification using backend
func (s *sender) Send(action string, data interface{}, level old_notifications.ImportanceLevel) error {
	template, ok := templates[action]
	if !ok {
		// skip sending if no template provided
		return nil
	}

	// perform data parsing and validation
	parser, ok := parsers[action]
	if !ok {
		return nil
	}
	phone, err := parser(data)
	if err != nil {
		return err
	}

	// render message
	body := format.Sprintf(template, data.(map[string]interface{}))

	// send it using backend
	return s.backend.Send(phone, body)
}

// templates
var templates = map[string]string{
	old_notifications.ActionRegistrationConfirmationRequested:     "Your ZamZam verification code - %<code>s",
	old_notifications.ActionPasswordRecoveryConfirmationRequested: "Your password recovery code - %<code>s",
}

//
func confirmationDataParser(data interface{}) (string, error) {
	m, ok := data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("expecting map[string]interface{} as data, not %T", data)
	}

	_, codeOk := m["code"]
	phone, phoneOk := m["phone"]
	if !codeOk || !phoneOk {
		return "", errors.New(`expecting both "code" and "phone" to be passed using data argument`)
	}
	return phone.(string), nil
}

// data parsers
var parsers = map[string]func(data interface{}) (string, error){
	old_notifications.ActionRegistrationConfirmationRequested:     confirmationDataParser,
	old_notifications.ActionPasswordRecoveryConfirmationRequested: confirmationDataParser,
}
