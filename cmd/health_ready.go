package cmd

import (
	"fmt"
	"os"

	"github.com/ory/oathkeeper/sdk/go/oathkeeper/client/api"
	"github.com/ory/x/cmdx"
	"github.com/spf13/cobra"
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
