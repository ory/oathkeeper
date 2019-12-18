package cmd

import (
	"fmt"
	"github.com/ory/oathkeeper/sdk/go/oathkeeper/client/api"
	"github.com/ory/x/cmdx"
	"github.com/spf13/cobra"
)

// healthCmd represents the health command
var healthAliveCmd = &cobra.Command{
	Use:   "alive",
	Short: "Command for checking alive status",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(cmd)

		r, err := client.API.IsInstanceAlive(api.NewIsInstanceAliveParams())
		cmdx.Must(err, "%s", err)
		fmt.Println(cmdx.FormatResponse(r.Payload))
	},
}

func init() {
	healthCmd.AddCommand(healthAliveCmd)
}
