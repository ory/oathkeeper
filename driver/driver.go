package driver

import (
	"github.com/ory/oathkeeper/driver/configuration"
)

type Driver interface {
	Configuration() configuration.Provider
	Registry() Registry
}
