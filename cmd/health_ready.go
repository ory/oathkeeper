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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ory/oathkeeper/internal/httpclient/client/api"
	"github.com/ory/x/cmdx"
)

// healthCmd represents the health command
var healthReadyCmd = &cobra.Command{
	Use:   "ready",
	Short: "Checks if an ORY Oathkeeper deployment is ready",
	Long: `Usage example:
  oathkeeper health --endpoint=http://localhost:4456/ ready

Note:
  The endpoint URL should point to a single ORY Oathkeeper deployment.
  If the endpoint URL points to a Load Balancer, these commands will effective test the Load Balancer.
`,
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(cmd)

		r, err := client.API.IsInstanceReady(api.NewIsInstanceReadyParams())
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
	healthCmd.AddCommand(healthReadyCmd)
}
