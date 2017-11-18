package cmd

import (
	"fmt"

	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/meatballhat/negroni-logrus"
	"github.com/ory/graceful"
	"github.com/ory/herodot"
	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/oathkeeper/rsakey"
	"github.com/ory/oathkeeper/rule"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
)

type managementConfig struct {
	hydra   *hydra.Configuration
	rules   rule.Manager
	address string
}

func runManagement(c *managementConfig) {
	sdk, err := hydra.NewSDK(c.hydra)
	if err != nil {
		logger.WithError(err).Fatalln("Unable to connect to Hydra SDK")
		return
	}

	keyManager := &rsakey.HydraManager{
		SDK: sdk,
		Set: viper.GetString("HYDRA_JWK_SET_ID"),
	}

	rules := rule.Handler{H: herodot.NewJSONWriter(logger), M: c.rules}
	keys := rsakey.Handler{H: herodot.NewJSONWriter(logger), M: keyManager}
	router := httprouter.New()
	rules.SetRoutes(router)
	keys.SetRoutes(router)

	n := negroni.New()
	n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oathkeeper-management"))
	n.UseHandler(router)

	go refreshKeys(keyManager, 0)

	addr := c.address
	server := graceful.WithDefaults(&http.Server{
		Addr:    addr,
		Handler: router,
	})

	logger.Printf("Listening on %s.\n", addr)
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		logger.Fatalf("Unable to gracefully shutdown HTTP server because %s.\n", err)
		return
	}
	logger.Println("HTTP server was shutdown gracefully")
}

// managementCmd represents the management command
var managementCmd = &cobra.Command{
	Use:   "management",
	Short: "Starts the ORY Oathkeeper management REST API",
	Long: `This starts a HTTP/2 REST API for managing ORY Oathkeeper.

CORE CONTROLS
=============

` + databaseUrl + `


HTTP CONTROLS
==============

- MANAGEMENT_HOST: The host to listen on.
	Default: PROXY_HOST="" (all interfaces)
- MANAGEMENT_PORT: The port to listen on.
	Default: PROXY_PORT="4456"


` + corsMessage,
	Run: func(cmd *cobra.Command, args []string) {
		rules, err := newRuleManager(viper.GetString("DATABASE_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to rule backend")
		}

		config := &managementConfig{
			hydra: &hydra.Configuration{
				ClientID:     viper.GetString("HYDRA_CLIENT_ID"),
				ClientSecret: viper.GetString("HYDRA_CLIENT_SECRET"),
				EndpointURL:  viper.GetString("HYDRA_URL"),
				Scopes:       []string{"hydra.warden"},
			},
			rules:   rules,
			address: fmt.Sprintf("%s:%s", viper.GetString("MANAGEMENT_HOST"), viper.GetString("MANAGEMENT_PORT")),
		}

		runManagement(config)
	},
}

func init() {
	serveCmd.AddCommand(managementCmd)
}
