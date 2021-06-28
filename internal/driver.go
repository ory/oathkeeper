// Copyright 2021 Ory GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
