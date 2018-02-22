// Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ory/oathkeeper/sdk/go/oathkeepersdk"
	"github.com/ory/oathkeeper/sdk/go/oathkeepersdk/swagger"
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import rules from a JSON file",
	Long: `The JSON file must be formatted as an array containing one or more rules:

[
	{ id: "rule-1", ... },
	{ id: "rule-2", ... },
]

Usage example:

	oathkeeper rules --endpoint=http://localhost:4456/ import rules.json
`,
	Run: func(cmd *cobra.Command, args []string) {
		endpoint, _ := cmd.Flags().GetString("endpoint")
		if endpoint == "" {
			fatalf("Please specify the endpoint url using the --endpoint flag, for more information use `oathkeeper help rules`")
		} else if len(args) != 1 {
			fatalf("Please specify a JSON file to load the rule definitions from, for more information use `oathkeeper help rules import`")
		}

		file, err := ioutil.ReadFile(args[0])
		must(err, "Reading file %s resulted in error %s", args[0], err)

		var rules []swagger.Rule
		err = json.Unmarshal(file, &rules)
		must(err, "Decoding file contents from JSON resulted in error %s", err)

		for _, r := range rules {
			fmt.Printf("Importing rule %s...\n", r.Id)
			client := oathkeepersdk.NewSDK(endpoint)
			out, response, err := client.CreateRule(r)
			checkResponse(response, err, http.StatusCreated)
			fmt.Printf("Successfully imported rule %s...\n", out.Id)
		}
		fmt.Printf("Successfully imported all rules from %s", args[0])
	},
}

func init() {
	rulesCmd.AddCommand(importCmd)
}
