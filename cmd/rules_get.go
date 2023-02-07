// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ory/oathkeeper/internal/httpclient/client/api"
	"github.com/ory/x/cmdx"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get access rule",
	Long: `Usage example:

	oathkeeper rules --endpoint=http://localhost:4456/ get rule-1
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdx.ExactArgs(cmd, args, 1)
		client := newClient(cmd)

		r, err := client.API.GetRule(api.NewGetRuleParams().WithID(args[0]))
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Could not get rule: %s", err)
			return cmdx.FailSilently(cmd)
		}
		fmt.Println(cmdx.FormatResponse(r.Payload))
		return nil
	},
}

func init() {
	rulesCmd.AddCommand(getCmd)
}
