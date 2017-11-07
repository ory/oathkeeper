package cmd

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/meatballhat/negroni-logrus"
	"github.com/ory/graceful"
	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/oathkeeper/director"
	"github.com/ory/oathkeeper/evaluator"
	"github.com/ory/oathkeeper/rule"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
)

type proxyConfig struct {
	hydra             *hydra.Configuration
	backendURL        string
	databaseURL       string
	cors              cors.Options
	address           string
	refreshDelay      string
	rules             rule.Manager
	bearerTokenSecret string
}

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Runs the reverse proxy",
	Run: func(cmd *cobra.Command, args []string) {
		rules, err := newRuleManager(viper.GetString("DATABASE_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to rule backend.")
		}

		config := &proxyConfig{
			hydra: &hydra.Configuration{
				ClientID:     viper.GetString("HYDRA_CLIENT_ID"),
				ClientSecret: viper.GetString("HYDRA_CLIENT_SECRET"),
				EndpointURL:  viper.GetString("HYDRA_URL"),
				Scopes:       []string{"hydra.warden"},
			},
			rules: rules, backendURL: viper.GetString("BACKEND_URL"),
			bearerTokenSecret: viper.GetString("JWT_SHARED_SECRET"),
			cors:              parseCorsOptions(""),
			address:           fmt.Sprintf("%s:%s", viper.GetString("PROXY_HOST"), viper.GetString("PROXY_PORT")),
			refreshDelay:      viper.GetString("REFRESH_DELAY"),
		}

		runProxy(config)
	},
}

func runProxy(c *proxyConfig) {
	sdk, err := hydra.NewSDK(c.hydra)
	if err != nil {
		logger.WithError(err).Fatalln("Unable to connect to Hydra SDK.")
		return
	}
	backend, err := url.Parse(c.backendURL)
	if err != nil {
		logger.WithError(err).Fatalln("Unable to parse backend URL.")
	}

	matcher := &rule.CachedMatcher{Manager: c.rules, Rules: []rule.Rule{}}

	if err := matcher.Refresh(); err != nil {
		logger.WithError(err).Fatalln("Unable to refresh rules.")
	}

	go refresh(c, matcher, 0)

	eval := evaluator.NewWardenEvaluator(logger, matcher, sdk)
	d := director.NewDirector(backend, eval, logger, c.bearerTokenSecret)
	proxy := &httputil.ReverseProxy{
		Director:  d.Director,
		Transport: d,
	}

	n := negroni.New()
	n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oahtkeeper-proxy"))
	n.UseHandler(proxy)

	ch := cors.New(c.cors).Handler(n)

	server := graceful.WithDefaults(&http.Server{
		Addr:    c.address,
		Handler: ch,
	})

	logger.Printf("Listening on %s.\n", c.address)
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		logger.Fatalf("Unable to gracefully shutdown HTTP server because %s.\n", err)
		return
	}
	logger.Println("HTTP server was shutdown gracefully")
}

func init() {
	serveCmd.AddCommand(proxyCmd)
}
