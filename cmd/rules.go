// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// rulesCmd represents the rules command
var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Commands for managing rules",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.UsageString())
	},
}

func init() {
	RootCmd.AddCommand(rulesCmd)
	rulesCmd.PersistentFlags().StringP("endpoint", "e", "", "The endpoint URL of ORY Oathkeeper's management API")
}
