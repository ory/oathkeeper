package cmd

import (
	"fmt"

	"github.com/ory/oathkeeper/sdk/go/oathkeeper/client/api"
	"github.com/ory/x/cmdx"
	"github.com/spf13/cobra"
)

// healthCmd represents the health command
var healthReadyCmd = &cobra.Command{
	Use:   "ready",
	Short: "Command for checking readiness status",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(cmd)

		r, err := client.API.IsInstanceReady(api.NewIsInstanceReadyParams())
		cmdx.Must(err, "%s", err)
		fmt.Println(cmdx.FormatResponse(r.Payload))
	},
}

func init() {
	healthCmd.AddCommand(healthReadyCmd)
}
