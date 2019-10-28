package x

import (
	"github.com/sirupsen/logrus"

	"github.com/ory/herodot"
)

type TestLoggerProvider struct{}

func (lp *TestLoggerProvider) Logger() logrus.FieldLogger {
	return logrus.New()
}

type RegistryLogger interface {
	Logger() logrus.FieldLogger
}

type RegistryWriter interface {
	Writer() herodot.Writer
}
