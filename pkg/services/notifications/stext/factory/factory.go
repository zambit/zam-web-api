package factory

import (
	"fmt"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications/stext"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications/stext/file"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications/stext/slack"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications/stext/twilio"
	"net/url"
	"strings"
)

// New creates backend guessing concrete type from URI, panic if guess is failed
func New(uri string) notifications.ISender {
	parsed, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	var backend stext.IBackend
	switch parsed.Scheme {
	case "https":
		switch {
		case strings.Contains(parsed.Host, "slack"):
			backend = slack.New(uri)
		case strings.Contains(parsed.Host, "api.twilio.com"):
			backend = twilio.New(uri)
		}
	case "file":
		backend = file.New(parsed.Path)
	}

	if backend == nil {
		panic(fmt.Errorf("unsupported simple-text backend specified with %s", uri))
	}

	return stext.New(backend)
}
