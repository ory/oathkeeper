package driver

import (
	"github.com/sirupsen/logrus"

	"github.com/ory/oathkeeper/driver/configuration"
)

type DefaultDriver struct {
	c configuration.Provider
	r Registry
}

func NewDefaultDriver(l logrus.FieldLogger, version, build, date string, validate bool) Driver {
	c := configuration.NewViperProvider(l)

	if validate {
		configuration.MustValidate(l, c)
	}

	r := NewRegistry(c).WithLogger(l).WithBuildInfo(version, build, date)

	return &DefaultDriver{r: r, c: c}
}

func (r *DefaultDriver) Configuration() configuration.Provider {
	return r.c
}

func (r *DefaultDriver) Registry() Registry {
	return r.r
}
