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
	"fmt"

	"os"

	"github.com/ory/oathkeeper/rule"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate <database-url>",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println(cmd.UsageString())
			return
		}

		db, err := connectToSql(args[0])
		if err != nil {
			fmt.Printf("Could not connect to database because %s.\n", err)
			os.Exit(1)
			return
		}

		m := rule.NewSQLManager(db)
		num, err := m.CreateSchemas()
		if err != nil {
			fmt.Printf("Could not create schemas because %s.\n", err)
			os.Exit(1)
			return
		}

		fmt.Printf("Successfully applied %d migrations.\n", num)
	},
}

func init() {
	RootCmd.AddCommand(migrateCmd)
}
