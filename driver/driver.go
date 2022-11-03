// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/ory/oathkeeper/driver/configuration"
)

type Driver interface {
	Configuration() configuration.Provider
	Registry() Registry
}
