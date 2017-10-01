package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"net/http/httputil"
	"github.com/ory/graceful"
	"net/http"
	"log"
	"github.com/ory/oathkeeper/director"
	"github.com/spf13/viper"
	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/oathkeeper/rule"
	"net/url"
)

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Runs the reverse proxy",
	Run: func(cmd *cobra.Command, args []string) {
		sdk, err := hydra.NewSDK(&hydra.Configuration{
			ClientID:     viper.GetString("HYDRA_CLIENT_ID"),
			ClientSecret: viper.GetString("HYDRA_CLIENT_SECRET"),
			EndpointURL:  viper.GetString("HYDRA_CLIENT_SECRET"),
			Scopes:       []string{"hydra.warden", "hydra.warden.*"},
		})
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to Hydra SDK.")
			return
		}

		builder := &rule.WardenRequestBuilder{

		}

		backend, err := url.Parse(viper.GetString("BACKEND_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to parse backend URL.")

		}

		d := director.NewDirector(backend, sdk, builder, logger, )
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
