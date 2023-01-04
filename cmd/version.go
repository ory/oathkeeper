// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ory/oathkeeper/x"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display this binary's version, build time and git hash of this build",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version:    %s\n", x.Version)
		fmt.Printf("Git Hash:   %s\n", x.Commit)
		fmt.Printf("Build Time: %s\n", x.Date)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
