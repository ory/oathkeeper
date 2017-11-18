package cmd

import (
	"fmt"

	"net/http"

	"github.com/ory/oathkeeper/sdk/go/oathkeepersdk"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a rule",
	Long: `Usage example:

	oathkeeper rules --endpoint=http://localhost:4456/ delete rule-1
`,
	Run: func(cmd *cobra.Command, args []string) {
		endpoint, _ := cmd.Flags().GetString("endpoint")
		if endpoint == "" {
			fatalf("Please specify the endpoint url using the --endpoint flag, for more information use `oathkeeper help rules`")
		} else if len(args) != 1 {
			fatalf("Please specify the rule id, for more information use `oathkeeper help rules delete`")
		}

		client := oathkeepersdk.NewSDK(endpoint)
		response, err := client.DeleteRule(args[0])
		checkResponse(response, err, http.StatusNoContent)
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
