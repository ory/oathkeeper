// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
	"github.com/ory/x/jwksx"
)

// credentialsGenerateCmd represents the generate command
var credentialsGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a key for the specified algorithm",
	Long: `Examples:

$ oathkeeper credentials generate --alg ES256 > jwks.json
$ oathkeeper credentials generate --alg RS256 > jwks.json
$ oathkeeper credentials generate --alg RS256 --bits 4096 > jwks.json`,
	Run: func(cmd *cobra.Command, _ []string) {
		key, err := jwksx.GenerateSigningKeys(
			flagx.MustGetString(cmd, "kid"),
			flagx.MustGetString(cmd, "alg"),
			flagx.MustGetInt(cmd, "bits"),
		)
		cmdx.Must(err, "Unable to generate key: %s", err)

		d := json.NewEncoder(os.Stdout)
		d.SetIndent("", "  ")
		err = d.Encode(key)
		cmdx.Must(err, "Unable to encode key to JSON: %s", err)
	},
}

func init() {
	credentialsCmd.AddCommand(credentialsGenerateCmd)

	credentialsGenerateCmd.Flags().String("alg", "", fmt.Sprintf("Generate a key to be used for one of the following algorithms: %v", jwksx.GenerateSigningKeysAvailableAlgorithms()))
	credentialsGenerateCmd.Flags().String("kid", "", "The JSON Web Key ID (kid) to be used. A random value will be used if left empty.")
	credentialsGenerateCmd.Flags().Int("bits", 0, "The key size in bits. If left empty will default to a secure value for the selected algorithm.")

	cmdx.Must(credentialsGenerateCmd.MarkFlagRequired("alg"), "")
}
