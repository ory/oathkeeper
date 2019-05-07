package internal

import (
	"github.com/spf13/viper"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/x/logrusx"
)

func ResetViper() {
	viper.Set(configuration.ViperKeyMutatorIDTokenJWKSURL, nil)

	viper.Set("LOG_LEVEL", "debug")
}

func NewConfigurationWithDefaults() *configuration.ViperProvider {
	ResetViper()
	return configuration.NewViperProvider(logrusx.New())
}

func NewRegistry(c *configuration.ViperProvider) *driver.RegistryMemory {
	viper.Set("LOG_LEVEL", "debug")
	r := driver.NewRegistryMemory().WithConfig(c)
	_ = r.Init()
	return r.(*driver.RegistryMemory)
}
