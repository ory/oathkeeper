// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"context"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
)

func NewConfigurationWithDefaults(opts ...configx.OptionModifier) configuration.Provider {
	l := logrusx.New("", "")
	c, err := configuration.NewKoanfProvider(
		context.Background(), nil, l, opts...)
	if err != nil {
		l.WithError(err).Fatal("Failed to initialize configuration")
	}
	return c
}

func NewRegistry(c configuration.Provider) *driver.RegistryMemory {
	return driver.NewRegistryMemory().WithConfig(c).(*driver.RegistryMemory)
}
