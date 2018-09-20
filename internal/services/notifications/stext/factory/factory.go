package factory

import (
	"fmt"
	old_notifications "git.zam.io/wallet-backend/web-api/internal/services/notifications"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications/stext"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications/file"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications/slack"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications/twilio"
	"net/url"
	"strings"
)

// New creates backend guessing concrete type from URI, panic if guess is failed
func New(uri string) (n old_notifications.ISender, err error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return
	}

	var transport notifications.ITransport
	switch parsed.Scheme {
	case "https":
		switch {
		case strings.Contains(parsed.Host, "slack"):
			transport = slack.New(uri)
		case strings.Contains(parsed.Host, "api.twilio.com"):
			transport, err = twilio.New(uri)
		}
	case "file":
		transport = file.New(parsed.Path)
	}

	if transport == nil {
		err = fmt.Errorf("unsupported simple-text transport specified with %s", uri)
	}

	if err == nil {
		n = stext.New(transport)
	}
	return
}
