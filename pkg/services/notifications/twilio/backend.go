package twilio

import (
	"encoding/json"
	"fmt"
	"regexp"

	"git.zam.io/wallet-backend/web-api/pkg/services/notifications"
	"github.com/pkg/errors"
	"github.com/ttacon/libphonenumber"

	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const errCodeAlphanumericFromNotAllowed = 21612

// backend just p
type backend struct {
	url, sid, token, fromPhone, fbFromPhone string
	regionFrom                              map[string]string
	client                                  http.Client
}

func getPhoneFromQuery(values url.Values, key, alterKey string) string {
	v := values.Get(key)
	if len(v) == 0 {
		v = values.Get(alterKey)
	}

	return strings.Replace(v, " ", "+", 1)
}

var fallbackRegionKeyRegExp = regexp.MustCompile("([a-zA-Z]{2})_from")

func getRegionSenders(q url.Values, ignoreSet map[string]struct{}) (r map[string]string, err error) {
	r = make(map[string]string)

	for k := range q {
		if _, ok := ignoreSet[k]; ok {
			continue
		}

		if match := fallbackRegionKeyRegExp.FindStringSubmatch(k); match != nil && len(match) == 2 {
			code := strings.ToUpper(match[1])
			_, isSupported := libphonenumber.GetSupportedRegions()[code]
			if !isSupported {
				return nil, fmt.Errorf(
					"twilio: region %s given with key %s are not supported", code, k,
				)
			}
			r[code] = getPhoneFromQuery(q, k, "")
		}
	}
	return
}

var ignoreQueryKeysSet = map[string]struct{}{
	"fallbackfrom":  {},
	"fallback_from": {},
	"from":          {},
}

// New creates new twillo transport form uri in format:
// 'https://{twilio_sid}:{twilio_token}@api.twilio.com/?From={send_from_phone}&FallbackFrom={fallback_send_from_phone}',
// where 'twilio_sid' and 'twilio_token' taken from your administrative console, 'send_from_phone' - is phone
// (both numeric and alphanumeric) from which messages will be sent, 'fallback_send_from_phone' - optional which will
// be used in case when recipient live in country where alphanumeric phone numbers are restricted
func New(uri string) (t notifications.ITransport, err error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return
	}
	qValues, err := url.ParseQuery(strings.ToLower(parsed.RawQuery))
	if err != nil {
		return
	}

	sid := parsed.User.Username()
	token, _ := parsed.User.Password()
	fromPhone := getPhoneFromQuery(qValues, "from", "")
	fbFromPhone := getPhoneFromQuery(qValues, "fallbackfrom", "fallback_from")

	tUrl := url.URL{
		Scheme:  parsed.Scheme,
		Host:    parsed.Host,
		RawPath: parsed.RawPath,
	}

	if sid == "" || token == "" || fromPhone == "" {
		err = fmt.Errorf(
			"error must match pattern: https://{twilio_sid}:{twilio_token}@api.twilio.com/?From={send_from_phone}&FallbackFrom={fallback_send_from_phone}",
		)
		return
	}

	// lookup region-specific senders
	s, err := getRegionSenders(qValues, ignoreQueryKeysSet)
	if err != nil {
		return
	}

	t = &backend{
		url:         tUrl.String() + "/2010-04-01/Accounts/" + sid + "/Messages.json",
		sid:         sid,
		token:       token,
		fromPhone:   fromPhone,
		fbFromPhone: fbFromPhone,
		regionFrom:  s,
	}
	return
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
func (b *backend) Send(recipient, body string) error {
	p, err := libphonenumber.Parse(recipient, "")
	if err != nil {
		return err
	}

	fromPhone := b.fromPhone
	region := libphonenumber.GetRegionCodeForNumber(p)
	if regionFrom, ok := b.regionFrom[region]; ok {
		fromPhone = regionFrom
	}

	err = b.send(fromPhone, recipient, body)
	if err == nil {
		// reverse err condition because it will simplify fallback flow
		return nil
	}

	if tErr, ok := err.(*respErr); ok {
		// in case when alphanumeric not allowed in recipient country, fallback on second from phone
		if tErr.Code == errCodeAlphanumericFromNotAllowed {
			if b.fbFromPhone != "" {
				return b.send(b.fbFromPhone, recipient, body)
			} else {
				return errors.Wrap(tErr, "alphanumeric phones not allowed in recipient country, but fallback phone not specified")
			}
		}
		return tErr
	}
	return err
}

func (b *backend) send(from, recipient, body string) error {
	// Build out the data for our message
	v := url.Values{}
	v.Set("To", recipient)
	v.Set("From", from)
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
