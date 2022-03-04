package driver

import (
	"context"

	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/driver/configuration"
)

type DefaultDriver struct {
	c configuration.Provider
	r Registry
}

func NewDefaultDriver(ctx context.Context, l *logrusx.Logger, version, build, date string) (Driver, error) {
	c, err := configuration.NewViperProvider(ctx, l)
	if err != nil {
		return nil, err
	}
	r := NewRegistry(c).WithLogger(l).WithBuildInfo(version, build, date)

	return &DefaultDriver{r: r, c: c}, nil
}

func (r *DefaultDriver) Configuration() configuration.Provider {
	return r.c
}

func (r *DefaultDriver) Registry() Registry {
	return r.r
}
