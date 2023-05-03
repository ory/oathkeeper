// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
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
		server.RunServe(x.Version, x.Commit, x.Date)(cmd, args)
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().Bool("disable-telemetry", false, "Disable anonymized telemetry reports - for more information please visit https://www.ory.sh/docs/ecosystem/sqa")
	serveCmd.PersistentFlags().Bool("sqa-opt-out", false, "Disable anonymized telemetry reports - for more information please visit https://www.ory.sh/docs/ecosystem/sqa")
}
