// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

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
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	configx.RegisterFlags(RootCmd.PersistentFlags())
}

func userHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}
