package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// allCmd represents the all command
var allCmd = &cobra.Command{
	Use:   "all",
	Short: "Runs both the proxy and management command",
	Long: `For documentation on the available configuration options please run:

* hydra help serve management
* hydra help serve proxy`,
	Run: func(cmd *cobra.Command, args []string) {
		rules, err := newRuleManager(viper.GetString("DATABASE_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to rule backend")
		}

		pc := &proxyConfig{rules: rules}
		mc := &managementConfig{rules: rules}

		go runManagement(mc)
		runProxy(pc)
	},
}

func init() {
	serveCmd.AddCommand(allCmd)
}
