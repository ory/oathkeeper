// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"

	_ "github.com/ory/jsonschema/v3/fileloader"
	_ "github.com/ory/jsonschema/v3/httploader"
	"github.com/ory/x/configx"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "oathkeeper",
	Short: "A cloud native Access and Identity Proxy",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		if !errors.Is(err, cmdx.ErrNoPrintButFail) {
			fmt.Println(err)
		}
		os.Exit(-1)
	}
}

func init() {
	configx.RegisterFlags(RootCmd.PersistentFlags())
}
