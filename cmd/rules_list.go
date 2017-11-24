package cmd

import (
	"net/http"

	"fmt"

	"github.com/ory/oathkeeper/sdk/go/oathkeepersdk"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available rules",
	Long: `Usage example:

	oathkeeper rules --endpoint=http://localhost:4456/ list
`,
	Run: func(cmd *cobra.Command, args []string) {
		endpoint, _ := cmd.Flags().GetString("endpoint")
		if endpoint == "" {
			fatalf("Please specify the endpoint url using the --endpoint flag, for more information use `oathkeeper help rules`")
		}

		client := oathkeepersdk.NewSDK(endpoint)
		rules, response, err := client.ListRules()
		checkResponse(response, err, http.StatusOK)
		fmt.Println(formatResponse(rules))
	},
}

func init() {
	rulesCmd.AddCommand(listCmd)
}
