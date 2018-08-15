package providers

import (
	"fmt"
	"git.zam.io/wallet-backend/web-api/config/server"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications/file"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications/slack"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications/twilio"
	"net/url"
	"strings"
)

// NotificationsTransport
func NotificationsTransport(cfg server.NotificatorScheme) (notifications.ITransport, error) {
	uri := cfg.URL
	parsed, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	var transport notifications.ITransport
	switch parsed.Scheme {
	case "https":
		switch {
		case strings.Contains(parsed.Host, "slack"):
			transport = slack.New(uri)
		case strings.Contains(parsed.Host, "api.twilio.com"):
			transport = twilio.New(uri)
		}
	case "file":
		transport = file.New(parsed.Path)
	}

	if transport == nil {
		return nil, fmt.Errorf("NotificationsTransport: failed to create transport due to wrong uri: %s", uri)
	}

	return transport, nil
}
