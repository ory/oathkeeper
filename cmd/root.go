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

	"github.com/gobuffalo/packr/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	_ "github.com/ory/jsonschema/v3/fileloader"
	_ "github.com/ory/jsonschema/v3/httploader"

	"github.com/ory/x/viperx"
)

var logger logrus.FieldLogger

var schemas = packr.New("schemas", "../.schema")

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

func watchAndValidateViper() {
	schema, err := schemas.Find("config.schema.json")
	if err != nil {
		logger.WithError(err).Fatal("Unable to open configuration JSON Schema.")
	}
	viperx.WatchAndValidateViper(logger, schema, "ORY Oathkeeper", []string{"serve", "profiling", "log"})
}

func init() {
	viperx.RegisterConfigFlag(RootCmd, "oathkeeper")
}
