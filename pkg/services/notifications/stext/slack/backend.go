package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

type requestBody struct {
	Text string `json:"text"`
}

// backend
type backend struct {
	client  http.Client
	hookUrl string
}

// New creates Slack backend using slack "Incoming WebHooks"
func New(hookUrl string) *backend {
	return &backend{hookUrl: hookUrl}
}

// Send notification using slack messages-hook
func (b *backend) Send(recipient string, body string) error {
	bodyDst := bytes.NewBuffer([]byte{})
	reqBody := requestBody{Text: fmt.Sprintf("Recipient %s\n%s", recipient, body)}
	encoder := json.NewEncoder(bodyDst)

	err := encoder.Encode(&reqBody)
	if err != nil {
		return err
	}

	resp, err := b.client.Post(b.hookUrl, "application/json", bodyDst)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		return nil
	}

	data := make([]byte, 500)
	read, err := io.ReadFull(resp.Body, data)
	if err != nil {
		return errors.Wrap(err, "occurred while reading bad slack response")
	}
	return fmt.Errorf("slack api response: %v", string(data[:read]))
}
