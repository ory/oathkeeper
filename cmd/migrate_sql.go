// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
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
	"github.com/spf13/viper"
)

// sqlCmd represents the sql command
var sqlCmd = &cobra.Command{
	Use: "sql <database-url>",
	Run: func(cmd *cobra.Command, args []string) {
		var dburl string
		if readFromEnv, _ := cmd.Flags().GetBool("read-from-env"); readFromEnv {
			if len(viper.GetString("DATABASE_URL")) == 0 {
				fmt.Println(cmd.UsageString())
				fmt.Println("")
				fmt.Println("When using flag -e, environment variable DATABASE_URL must be set")
				return
			}
			dburl = viper.GetString("DATABASE_URL")
		} else {
			if len(args) != 1 {
				fmt.Println(cmd.UsageString())
				return
			}
			dburl = args[0]
		}

		db, err := connectToSql(dburl)
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
	migrateCmd.AddCommand(sqlCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sqlCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sqlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	sqlCmd.Flags().BoolP("read-from-env", "e", false, "If set, reads the database URL from the environment variable DATABASE_URL.")
}
