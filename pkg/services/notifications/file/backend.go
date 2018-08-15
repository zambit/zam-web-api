package file

import (
	"bytes"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications"
	"os"
	"time"
)

// transport
type transport struct {
	file *os.File
}

// New
func New(fileName string) notifications.ITransport {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		panic(err)
	}

	return transport{file: f}
}

// Send
func (b transport) Send(recipient string, body string) error {
	buf := bytes.Buffer{}
	buf.WriteString(time.Now().UTC().Format(time.UnixDate))
	buf.WriteString(" ")
	buf.WriteString(recipient)
	buf.WriteString(" - ")
	buf.WriteString(body)
	buf.WriteString("\n")
	_, err := buf.WriteTo(b.file)
	return err
}
