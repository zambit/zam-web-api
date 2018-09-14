package providers

import (
	"git.zam.io/wallet-backend/web-api/config/logging"
	"github.com/sirupsen/logrus"
)

// RootLogger
func RootLogger(cfg logging.Scheme) logrus.FieldLogger {
	l := logrus.New()

	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}

	f := logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	}
	l.SetFormatter(&f)

	l.SetLevel(level)
	return l
}
