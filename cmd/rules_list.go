// Copyright 2021 Ory GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
