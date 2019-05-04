/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ory/oathkeeper/sdk/go/oathkeeper/client/rule"
	"github.com/ory/x/cmdx"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a rule",
	Long: `Usage example:

	oathkeeper rules --endpoint=http://localhost:4456/ delete rule-1
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fatalf("Please specify the rule id, for more information use `oathkeeper help rules delete`")
		}

		client := newClient(cmd)
		_, err := client.Rule.DeleteRule(rule.NewDeleteRuleParams().WithID(args[0]))
		cmdx.Must(err, "%s", err)
		fmt.Printf("Successfully deleted rule %s\n", args[0])
	},
}

func init() {
	rulesCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
