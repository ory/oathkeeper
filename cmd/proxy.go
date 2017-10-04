package cmd

import (
	"fmt"

	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/ory/graceful"
	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/oathkeeper/director"
	"github.com/ory/oathkeeper/evaluator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Runs the reverse proxy",
	Run: func(cmd *cobra.Command, args []string) {
		sdk, err := hydra.NewSDK(&hydra.Configuration{
			ClientID:     viper.GetString("HYDRA_CLIENT_ID"),
			ClientSecret: viper.GetString("HYDRA_CLIENT_SECRET"),
			EndpointURL:  viper.GetString("HYDRA_ENDPOINT_URL"),
			Scopes:       []string{"hydra.warden", "hydra.warden.*"},
		})
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to Hydra SDK.")
			return
		}
		backend, err := url.Parse(viper.GetString("BACKEND_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to parse backend URL.")

		}

		eval := evaluator.NewWardenEvaluator(logger, nil, sdk)
		d := director.NewDirector(backend, eval, logger, viper.GetString("JWT_SECRET"))
		proxy := &httputil.ReverseProxy{
			Director:  d.Director,
			Transport: d,
		}

		server := graceful.WithDefaults(&http.Server{
			Addr:    fmt.Sprintf("%s:%s", viper.GetString("PROXY_HOST"), viper.GetString("PROXY_PORT")),
			Handler: proxy,
		})

		if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
			log.Fatalf("Unable to gracefully shutdown HTTP server becase %s.\n", err)
			return
		}
		log.Println("HTTP server was shutdown gracefully")
	},
}

func init() {
	RootCmd.AddCommand(proxyCmd)
}
