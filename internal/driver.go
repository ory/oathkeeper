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

	// We need to reset the default value as defined in configuration.init()
	viper.SetDefault(configuration.ViperKeyErrorsJSONIsEnabled, true)
}

func NewConfigurationWithDefaults() *configuration.ViperProvider {
	ResetViper()
	return configuration.NewViperProvider(logrusx.New("", ""))
}

func NewRegistry(c *configuration.ViperProvider) *driver.RegistryMemory {
	viper.Set("LOG_LEVEL", "debug")
	return driver.NewRegistryMemory().WithConfig(c).(*driver.RegistryMemory)
}
