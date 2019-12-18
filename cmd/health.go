package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Commands for checking health",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.UsageString())
	},
}

func init() {
	RootCmd.AddCommand(healthCmd)
	healthCmd.PersistentFlags().StringP("endpoint", "e", "", "The endpoint URL of ORY Oathkeeper's management API")
}
