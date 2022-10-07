// Copyright © 2022 Ory Corp

package cmd

import (
	"github.com/spf13/cobra"
)

// credentialsCmd represents the credentials command
var credentialsCmd = &cobra.Command{
	Use:   "credentials",
	Short: "Generate RSA, ECDSA, and other keys and output them as JSON Web Keys",
}

func init() {
	RootCmd.AddCommand(credentialsCmd)
}
