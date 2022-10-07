// Copyright © 2022 Ory Corp

package driver

import (
	"github.com/ory/oathkeeper/driver/configuration"
)

type Driver interface {
	Configuration() configuration.Provider
	Registry() Registry
}
