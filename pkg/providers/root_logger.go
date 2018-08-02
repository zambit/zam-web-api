package providers

import (
	"github.com/sirupsen/logrus"
)

// RootLogger
func RootLogger() logrus.FieldLogger {
	return logrus.New()
}
