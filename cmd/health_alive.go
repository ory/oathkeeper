package cmd

import (
	"fmt"
	"os"

	"github.com/ory/oathkeeper/sdk/go/oathkeeper/client/api"
	"github.com/ory/x/cmdx"
	"github.com/spf13/cobra"
)

// healthCmd represents the health command
var healthAliveCmd = &cobra.Command{
	Use:   "alive",
	Short: "Checks if an ORY Oathkeeper deployment is alive",
	Long: `Usage example:
  oathkeeper health --endpoint=http://localhost:4456/ alive
`,
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(cmd)

		r, err := client.API.IsInstanceAlive(api.NewIsInstanceAliveParams())
		// If err, print err and exit 1
		cmdx.Must(err, "%s", err)
		// Print payload
		fmt.Println(cmdx.FormatResponse(r.Payload))
		// When healthy, ORY Oathkeeper always returns a status of "ok"
		if r.Payload.Status != "ok" {
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	healthCmd.AddCommand(healthAliveCmd)
}
