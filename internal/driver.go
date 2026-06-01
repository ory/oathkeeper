// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
)

func NewConfigurationWithDefaults(t testing.TB, opts ...configx.OptionModifier) configuration.Configuration {
	l := logrusx.NewT(t)
	c, err := configuration.NewKoanfProvider(t.Context(), nil, l, opts...)
	require.NoError(t, err)
	return c
}

func NewRegistry(t testing.TB, opts ...configx.OptionModifier) *driver.RegistryMemory {
	l := logrusx.NewT(t)
	c, err := configuration.NewKoanfProvider(t.Context(), nil, l, opts...)
	require.NoError(t, err)
	return driver.NewRegistryMemory(c, l)
}
