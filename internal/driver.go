package internal

import (
	"github.com/ory/viper"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/x/logrusx"
)

func ResetViper() {
	viper.Reset()
	viper.Set("log.level", "debug")
}

func NewConfigurationWithDefaults() *configuration.ViperProvider {
	ResetViper()
	return configuration.NewViperProvider(logrusx.New())
}

func NewRegistry(c *configuration.ViperProvider) *driver.RegistryMemory {
	viper.Set("LOG_LEVEL", "debug")
	return driver.NewRegistryMemory().WithConfig(c).(*driver.RegistryMemory)
}
