/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package cmd

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/gobuffalo/packr/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	_ "github.com/ory/jsonschema/v3/fileloader"
	_ "github.com/ory/jsonschema/v3/httploader"

	"github.com/ory/viper"
	"github.com/ory/x/viperx"
)

var logger logrus.FieldLogger

var schemas = packr.New("schemas", "../.schemas")

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "oathkeeper",
	Short: "A cloud native Access and Identity Proxy",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	viperx.RegisterConfigFlag(RootCmd, "oathkeeper")
}

func watchAndValidateViper() {
	logger = viperx.InitializeConfig("oathkeeper", "", logger)

	schema, err := schemas.Find("config.schema.json")
	if err != nil {
		logger.WithError(err).Fatal("Unable to open configuration JSON Schema.")
	}

	if err := viperx.Validate("config.schema.json", schema); err != nil {
		viperx.LoggerWithValidationErrorFields(logger, err).
			Fatal("The configuration is invalid and could not be loaded.")
	}

	viperx.AddWatcher(func(event fsnotify.Event) error {
		if err := viperx.Validate("config.schema.json", schema); err != nil {
			viperx.LoggerWithValidationErrorFields(logger, err).
				Error("The changed configuration is invalid and could not be loaded. Rolling back to the last working configuration revision. Please address the validation errors before restarting ORY Oathkeeper.")
			return viperx.ErrRollbackConfigurationChanges
		}
		return nil
	})

	viperx.WatchConfig(logger, &viperx.WatchOptions{
		Immutables: []string{"serve", "profiling", "log"},
		OnImmutableChange: func(key string) {
			logger.
				WithField("key", key).
				WithField("reset_to", fmt.Sprintf("%v", viper.Get(key))).
				Error("A configuration value marked as immutable has changed. Rolling back to the last working configuration revision. To reload the values please restart ORY Oathkeeper.")
		},
	})
}
