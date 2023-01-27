// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/ory/x/flagx"

	"github.com/spf13/cobra"

	"github.com/ory/oathkeeper/internal/httpclient/client/api"
	"github.com/ory/x/cmdx"
)

// rulesListCmd represents the list command
var rulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List access rules",
	Long: `Usage example:

	oathkeeper rules --endpoint=http://localhost:4456/ list
`,
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient(cmd)

		limit, page := int64(flagx.MustGetInt(cmd, "limit")), int64(flagx.MustGetInt(cmd, "page"))
		offset := (limit * page) - limit

		r, err := client.API.ListRules(api.NewListRulesParams().WithLimit(&limit).WithOffset(&offset))
		cmdx.Must(err, "%s", err)
		fmt.Println(cmdx.FormatResponse(r.Payload))
	},
}

func init() {
	rulesCmd.AddCommand(rulesListCmd)
	rulesListCmd.Flags().Int("limit", 20, "The maximum amount of policies returned.")
	rulesListCmd.Flags().Int("page", 1, "The number of page.")
}
