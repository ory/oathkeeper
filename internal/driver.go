package internal

import (
	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/x/logrusx"
	"github.com/spf13/viper"
)

func resetConfig() {
	viper.Set(configuration.ViperKeyMutatorIDTokenJWKSURL, nil)

	viper.Set("LOG_LEVEL", "debug")
}

func NewConfigurationWithDefaults() *configuration.ViperProvider {
	resetConfig()
	return configuration.NewViperProvider(logrusx.New())
}

func NewRegistry(c *configuration.ViperProvider) *driver.RegistryMemory {
	viper.Set("LOG_LEVEL", "debug")
	r := driver.NewRegistryMemory().WithConfig(c)
	_ = r.Init()
	return r.(*driver.RegistryMemory)
}
