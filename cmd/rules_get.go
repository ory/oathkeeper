package cmd

import (
	"net/http"

	"fmt"

	"github.com/ory/oathkeeper/sdk/go/oathkeepersdk"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Fetch a rule",
	Long: `Usage example:

	oathkeeper rules --endpoint=http://localhost:4456/ get rule-1
`,
	Run: func(cmd *cobra.Command, args []string) {
		endpoint, _ := cmd.Flags().GetString("endpoint")
		if endpoint == "" {
			fatalf("Please specify the endpoint url using the --endpoint flag, for more information use `oathkeeper help rules`")
		} else if len(args) != 1 {
			fatalf("Please specify the rule id, for more information use `oathkeeper help rules get`")
		}

		client := oathkeepersdk.NewSDK(endpoint)
		rule, response, err := client.GetRule(args[0])
		checkResponse(response, err, http.StatusOK)
		fmt.Println(formatResponse(rule))
	},
}

func init() {
	rulesCmd.AddCommand(getCmd)
}
