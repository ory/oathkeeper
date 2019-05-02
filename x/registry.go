package x

import (
	"github.com/sirupsen/logrus"

	"github.com/ory/herodot"
)

type RegistryLogger interface {
	Logger() logrus.FieldLogger
}

type RegistryWriter interface {
	Writer() herodot.Writer
}
