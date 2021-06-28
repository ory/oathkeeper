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

	"github.com/spf13/cobra"
)

// healthCmd represents the health command
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Commands for checking the status of an ORY Oathkeeper deployment",
	Long: `Note:
  The endpoint URL should point to a single ORY Oathkeeper deployment.
  If the endpoint URL points to a Load Balancer, these commands will effective test the Load Balancer.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.UsageString())
	},
}

func init() {
	RootCmd.AddCommand(healthCmd)
	healthCmd.PersistentFlags().StringP("endpoint", "e", "", "The endpoint URL of ORY Oathkeeper's management API")
}
