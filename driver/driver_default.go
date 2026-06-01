// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"fmt"

	"github.com/spf13/pflag"

	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/driver/configuration"
)

func NewDefaultDriver(ctx context.Context, l *logrusx.Logger, flags *pflag.FlagSet, configOpts ...configx.OptionModifier) (Registry, error) {
	c, err := configuration.NewKoanfProvider(ctx, flags, l, configOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize configuration provider: %w", err)
	}
	return NewRegistry(c, l), nil
}
