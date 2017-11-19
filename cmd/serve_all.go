package cmd

import (
	"fmt"

	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// allCmd represents the all command
var allCmd = &cobra.Command{
	Use:   "all",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		rules, err := newRuleManager(viper.GetString("DATABASE_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to rule backend")
		}

		pc := &proxyConfig{
			hydra: &hydra.Configuration{
				ClientID:     viper.GetString("HYDRA_CLIENT_ID"),
				ClientSecret: viper.GetString("HYDRA_CLIENT_SECRET"),
				EndpointURL:  viper.GetString("HYDRA_URL"),
				Scopes:       []string{"hydra.warden", "hydra.keys.*"},
			},
			rules: rules, backendURL: viper.GetString("BACKEND_URL"),
			cors:         parseCorsOptions(""),
			address:      fmt.Sprintf("%s:%s", viper.GetString("PROXY_HOST"), viper.GetString("PROXY_PORT")),
			refreshDelay: viper.GetString("RULES_REFRESH_INTERVAL"),
		}

		mc := &managementConfig{
			hydra: &hydra.Configuration{
				ClientID:     viper.GetString("HYDRA_CLIENT_ID"),
				ClientSecret: viper.GetString("HYDRA_CLIENT_SECRET"),
				EndpointURL:  viper.GetString("HYDRA_URL"),
				Scopes:       []string{"hydra.warden", "hydra.keys.*"},
			},
			rules:   rules,
			address: fmt.Sprintf("%s:%s", viper.GetString("MANAGEMENT_HOST"), viper.GetString("MANAGEMENT_PORT")),
		}

		go runManagement(mc)
		runProxy(pc)
	},
}

func init() {
	serveCmd.AddCommand(allCmd)
}
