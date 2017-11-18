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
