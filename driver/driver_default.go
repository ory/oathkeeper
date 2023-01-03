// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"

	"github.com/spf13/pflag"

	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/driver/configuration"
)

type DefaultDriver struct {
	c configuration.Provider
	r Registry
}

func NewDefaultDriver(l *logrusx.Logger, version, build, date string, flags *pflag.FlagSet, configOpts ...configx.OptionModifier) Driver {
	c, err := configuration.NewKoanfProvider(
		context.Background(), flags, l, configOpts...)
	if err != nil {
		l.WithError(err).Fatal("Failed to initialize configuration")
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
