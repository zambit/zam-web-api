package twilio

import (
	"encoding/json"
	"fmt"
	"git.zam.io/wallet-backend/web-api/services/notifications/stext"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// backend just p
type backend struct {
	url, sid, token, fromPhone string

	client http.Client
}

// New creates new twillo backend in form https://{twilio_sid}:{twilio_token}@api.twilio.com/?From={send_from_phone}
func New(uri string) stext.IBackend {
	parsed, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	sid := parsed.User.Username()
	token, _ := parsed.User.Password()
	fromPhone := parsed.Query().Get("From")
	fromPhone = strings.Replace(fromPhone, " ", "+", 1)

	tUrl := url.URL{
		Scheme:  parsed.Scheme,
		Host:    parsed.Host,
		RawPath: parsed.RawPath,
	}

	if sid == "" || token == "" || fromPhone == "" {
		panic(fmt.Errorf(
			"error must match pattern: https://{twilio_sid}:{twilio_token}@api.twilio.com/?From={send_from_phone}",
		))
	}

	return &backend{
		url:       tUrl.String() + "/2010-04-01/Accounts/" + sid + "/Messages.json",
		sid:       sid,
		token:     token,
		fromPhone: fromPhone,
	}
}

// respErr used ta parse twillo error response at same time implements error
type respErr struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	MoreInfoUrl string `json:"more_info"`
}

// Error
func (e *respErr) Error() string {
	return fmt.Sprintf("%d: %s (%s)", e.Code, e.Message, e.MoreInfoUrl)
}

// Send
func (b *backend) Send(recipient string, body string) error {
	// Build out the data for our message
	v := url.Values{}
	v.Set("To", recipient)
	v.Set("From", b.fromPhone)
	v.Set("Body", body)
	rb := *strings.NewReader(v.Encode())

	req, _ := http.NewRequest("POST", b.url, &rb)
	req.SetBasicAuth(b.sid, b.token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Make request
	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// noting to do with response body
		return nil
	}

	// parse response body
	e := &respErr{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bodyBytes, e)
	if err != nil {
		return err
	}
	return e
}
