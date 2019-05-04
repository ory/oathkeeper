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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	swaggerRule "github.com/ory/oathkeeper/sdk/go/oathkeeper/client/rule"
	"github.com/ory/oathkeeper/sdk/go/oathkeeper/models"
	"github.com/ory/x/cmdx"

	"github.com/spf13/cobra"

	"github.com/ory/oathkeeper/rule"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Imports rules from a JSON file",
	Long: `Imported rules are either created or updated if they already exist.

The JSON file must be formatted as an array containing one or more rules:

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

		var rules []rule.Rule
		d := json.NewDecoder(bytes.NewBuffer(file))
		d.DisallowUnknownFields()
		err = d.Decode(&rules)
		must(err, "Decoding file contents from JSON resulted in error %s", err)

		for _, r := range rules {
			fmt.Printf("Importing rule %s...\n", r.ID)
			client := newClient(cmd)

			shouldUpdate := false
			if _, err := client.Rule.GetRule(swaggerRule.NewGetRuleParams().WithID(r.ID)); err == nil {
				shouldUpdate = true
			}

			rh := make([]*models.SwaggerRuleHandler, len(r.Authenticators))
			for k, authn := range r.Authenticators {
				rh[k] = &models.SwaggerRuleHandler{
					Handler: authn.Handler,
					Config:  json.RawMessage(authn.Config),
				}
			}

			sr := models.SwaggerRule{
				ID:          r.ID,
				Description: r.Description,
				Match:       &models.SwaggerRuleMatch{Methods: r.Match.Methods, URL: r.Match.URL},
				Authorizer: &models.SwaggerRuleHandler{
					Handler: r.Authorizer.Handler,
					Config:  models.RawMessage(r.Authorizer.Config),
				},
				Authenticators: rh,
				CredentialsIssuer: &models.SwaggerRuleHandler{
					Handler: r.CredentialsIssuer.Handler,
					Config:  models.RawMessage(r.CredentialsIssuer.Config),
				},
				Upstream: &models.Upstream{
					URL:          r.Upstream.URL,
					PreserveHost: r.Upstream.PreserveHost,
					StripPath:    r.Upstream.StripPath,
				},
			}

			if shouldUpdate {
				response, err := client.Rule.UpdateRule(swaggerRule.NewUpdateRuleParams().WithID(r.ID).WithBody(&sr))
				cmdx.Must(err, "%s", err)
				fmt.Printf("Successfully imported rule %s...\n", response.Payload.ID)
			} else {
				response, err := client.Rule.CreateRule(swaggerRule.NewCreateRuleParams().WithBody(&sr))
				cmdx.Must(err, "%s", err)
				fmt.Printf("Successfully imported rule %s...\n", response.Payload.ID)
			}
		}
		fmt.Printf("Successfully imported all rules from %s", args[0])
	},
}

func init() {
	rulesCmd.AddCommand(importCmd)
}
