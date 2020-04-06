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
	"github.com/ory/x/logrusx"
	"github.com/ory/x/viperx"

	"github.com/spf13/cobra"

	"github.com/ory/oathkeeper/cmd/server"
	"github.com/ory/oathkeeper/x"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the HTTP/2 REST API and HTTP/2 Reverse Proxy",
	Long: `Opens two ports for serving both the HTTP/2 Rest API and the HTTP/2 Reverse Proxy.

## Configuration

ORY Oathkeeper can be configured using environment variables as well as a configuration file. For more information
on configuration options, open the configuration documentation:

>> https://www.ory.sh/oathkeeper/docs/configuration <<
`,
	Run: func(cmd *cobra.Command, args []string) {
		logger = viperx.InitializeConfig("oathkeeper", "", logger)

		watchAndValidateViper()
		server.RunServe(x.Version, x.Commit, x.Date)(cmd, args)
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)

	disableTelemetryEnv := viperx.GetBool(logrusx.New(), "sqa.opt_out", false, "DISABLE_TELEMETRY")
	serveCmd.PersistentFlags().Bool("disable-telemetry", disableTelemetryEnv, "Disable anonymized telemetry reports - for more information please visit https://www.ory.sh/docs/ecosystem/sqa")
	serveCmd.PersistentFlags().Bool("sqa-opt-out", disableTelemetryEnv, "Disable anonymized telemetry reports - for more information please visit https://www.ory.sh/docs/ecosystem/sqa")
}
