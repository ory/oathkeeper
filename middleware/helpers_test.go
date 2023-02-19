// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
)

// WithRegistry is a test option that writes the internal registry used in the
// middleware to the given pointer. The pointer must be allocaded by the caller
// beforehand (e.g., `regPtr := new(driver.Registry)`).
func WithRegistry(r *driver.Registry) Option {
	return func(o *options) { o.registryAddr = r }
}

// WithConfigProvider is a test option that writes the internal config provider used in the
// middleware to the given pointer. The pointer must be allocaded by the caller
// beforehand (e.g., `cfgPtr := new(configuration.Provider)`).
func WithConfigProvider(c *configuration.Provider) Option {
	return func(o *options) { o.configProviderAddr = c }
}
