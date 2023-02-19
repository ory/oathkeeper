// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/ory/oathkeeper/internal/httpclient/client/health"

	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"
)

// healthCmd represents the health command
var healthAliveCmd = &cobra.Command{
	Use:   "alive",
	Short: "Checks if an ORY Oathkeeper deployment is alive",
	Long: `Usage example:
  oathkeeper health --endpoint=http://localhost:4456/ alive

Note:
  The endpoint URL should point to a single ORY Oathkeeper deployment.
  If the endpoint URL points to a Load Balancer, these commands will effective test the Load Balancer.
`,
	Run: func(cmd *cobra.Command, _ []string) {
		client := newClient(cmd)

		r, err := client.Health.IsInstanceAlive(health.NewIsInstanceAliveParams())
		// If err, print err and exit 1
		cmdx.Must(err, "%s", err)
		// Print payload
		fmt.Println(cmdx.FormatResponse(r.Payload))
		// When healthy, ORY Oathkeeper always returns a status of "ok"
		if r.Payload.Status != "ok" {
			os.Exit(1)
		}
	},
}

func init() {
	healthCmd.AddCommand(healthAliveCmd)
}
