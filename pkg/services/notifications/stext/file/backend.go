package file

import (
	"bytes"
	"git.zam.io/wallet-backend/web-api/pkg/services/notifications/stext"
	"os"
	"time"
)

// backend
type backend struct {
	file *os.File
}

// New
func New(fileName string) stext.IBackend {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		panic(err)
	}

	return backend{file: f}
}

// Send
func (b backend) Send(recipient string, body string) error {
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
