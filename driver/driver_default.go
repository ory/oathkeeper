package driver

import (
	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/driver/configuration"
)

type DefaultDriver struct {
	c configuration.Provider
	r Registry
}

func NewDefaultDriver(l *logrusx.Logger, version, build, date string) Driver {
	c := configuration.NewViperProvider(l)
	r := NewRegistry(c).WithLogger(l).WithBuildInfo(version, build, date)

	return &DefaultDriver{r: r, c: c}
}

func (r *DefaultDriver) Configuration() configuration.Provider {
	return r.c
}

func (r *DefaultDriver) Registry() Registry {
	return r.r
}
