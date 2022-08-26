package configuration

import (
	"context"

	"github.com/ory/x/logrusx"
)

func NewViperProvider(l *logrusx.Logger) Provider {
	c, err := NewKoanfProvider(
		context.Background(), nil, l)
	if err != nil {
		l.WithError(err).Fatal("Failed to initialize configuration")
	}
	return c
}
