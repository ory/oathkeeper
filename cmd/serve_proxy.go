package cmd

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"crypto/tls"

	"encoding/base64"

	"github.com/meatballhat/negroni-logrus"
	"github.com/ory/graceful"
	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/oathkeeper/director"
	"github.com/ory/oathkeeper/evaluator"
	"github.com/ory/oathkeeper/rsakey"
	"github.com/ory/oathkeeper/rule"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
)

type proxyConfig struct {
	hydra        *hydra.Configuration
	backendURL   string
	databaseURL  string
	cors         cors.Options
	address      string
	refreshDelay string
	rules        rule.Manager
	tlsCert      string
	tlsKey       string
}

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Starts the ORY Oathkeeper firewall reverse proxy",
	Long: `This starts a HTTP/2 reverse proxy capable of deciding whether to forward API requests or to block them based on a set of rules.

This command exposes a variety of controls via environment variables. You can
set environments using "export KEY=VALUE" (Linux/macOS) or "set KEY=VALUE" (Windows). On Linux,
you can also set environments by prepending key value pairs: "KEY=VALUE KEY2=VALUE2 hydra"

All possible controls are listed below.

REQUIRED CONTROLS
=============

` + databaseUrl + `

- HYDRA_CLIENT_ID: The OAuth 2.0 Client ID to be used to connect to ORY Hydra. The client must allowed to request the
	hydra.warden OAuth 2.0 Scope and allowed to access the warden resources.

- HYDRA_CLIENT_SECRET: The OAuth 2.0 Client Secret of the Client ID referenced aboce.

- HYDRA_URL: The URL of ORY Hydra.
	Example: HYDRA_URL=https://hydra.com/

- BACKEND_URL: The URL where requests will be forwarded to, if access is granted.
	Example: BACKEND_URL=https://my-backend.com/

- JWT_SHARED_SECRET: The shared secret to be used to encrypt the Authorization Bearer JSON Web Token. Use this
	secret to validate that the Bearer Token was indeed issued by this ORY Oathkeeper instance.


HTTP(S) CONTROLS
==============

- HTTP_TLS_KEY: Base64 encoded (without padding) string of the private key (PEM encoded) to be used for HTTP over TLS (HTTPS).
	If not set, HTTPS will be disabled and instead HTTP will be served.

- HTTP_TLS_CERT: Base64 encoded (without padding) string of the TLS certificate (PEM encoded) to be used for HTTP over TLS (HTTPS).
	If not set, HTTPS will be disabled and instead HTTP will be served.

- PROXY_HOST: The host to listen on.
	Default: PROXY_HOST="" (all interfaces)

- PROXY_PORT: The port to listen on.
	Default: PROXY_PORT="4455"


OTHER CONTROLS
==============
- RULES_REFRESH_INTERVAL: ORY Oathkeeper stores rules in memory for faster access. This value sets the database polling interval.
	Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Default: RULES_REFRESH_INTERVAL=5s

` + corsMessage,
	Run: func(cmd *cobra.Command, args []string) {
		rules, err := newRuleManager(viper.GetString("DATABASE_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to rule backend")
		}

		config := &proxyConfig{
			hydra: &hydra.Configuration{
				ClientID:     viper.GetString("HYDRA_CLIENT_ID"),
				ClientSecret: viper.GetString("HYDRA_CLIENT_SECRET"),
				EndpointURL:  viper.GetString("HYDRA_URL"),
				Scopes:       []string{"hydra.warden"},
			},
			rules: rules, backendURL: viper.GetString("BACKEND_URL"),
			cors:         parseCorsOptions(""),
			address:      fmt.Sprintf("%s:%s", viper.GetString("PROXY_HOST"), viper.GetString("PROXY_PORT")),
			refreshDelay: viper.GetString("RULES_REFRESH_INTERVAL"),
			tlsKey:       viper.GetString("HTTP_TLS_KEY"),
			tlsCert:      viper.GetString("HTTP_TLS_CERT"),
		}

		runProxy(config)
	},
}

func runProxy(c *proxyConfig) {
	sdk, err := hydra.NewSDK(c.hydra)
	if err != nil {
		logger.WithError(err).Fatalln("Unable to connect to Hydra SDK")
		return
	}
	backend, err := url.Parse(c.backendURL)
	if err != nil {
		logger.WithError(err).Fatalln("Unable to parse backend URL")
	}

	matcher := &rule.CachedMatcher{Manager: c.rules, Rules: []rule.Rule{}}

	if err := matcher.Refresh(); err != nil {
		logger.WithError(err).Fatalln("Unable to refresh rules")
	}

	keyManager := &rsakey.HydraManager{
		SDK: sdk,
		Set: viper.GetString("HYDRA_JWK_SET_ID"),
	}

	go refreshRules(c, matcher, 0)
	go refreshKeys(keyManager, 0)

	eval := evaluator.NewWardenEvaluator(logger, matcher, sdk)
	d := director.NewDirector(backend, eval, logger, keyManager)
	proxy := &httputil.ReverseProxy{
		Director:  d.Director,
		Transport: d,
	}

	n := negroni.New()
	n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oahtkeeper-proxy"))
	n.UseHandler(proxy)

	ch := cors.New(c.cors).Handler(n)

	var cert tls.Certificate
	if c.tlsCert != "" && c.tlsKey != "" {
		if tlsCert, err := base64.StdEncoding.DecodeString(c.tlsCert); err != nil {
			logger.WithError(err).Fatalln("Unable to base64 decode the TLS Certificate")
		} else if tlsKey, err := base64.StdEncoding.DecodeString(c.tlsKey); err != nil {
			logger.WithError(err).Fatalln("Unable to base64 decode the TLS Private Key")
		} else if cert, err = tls.X509KeyPair(tlsCert, tlsKey); err != nil {
			logger.WithError(err).Fatalln("Unable to load X509 key pair")
		}
	}

	server := graceful.WithDefaults(&http.Server{
		Addr:    c.address,
		Handler: ch,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
		},
	})

	logger.Printf("Listening on %s.\n", c.address)
	if err := graceful.Graceful(func() error {
		if c.tlsCert != "" && c.tlsKey != "" {
			return server.ListenAndServeTLS("", "")
		}
		return server.ListenAndServe()
	}, server.Shutdown); err != nil {
		logger.Fatalf("Unable to gracefully shutdown HTTP(s) server because %s.\n", err)
		return
	}
	logger.Println("HTTP(s) server was shutdown gracefully")
}

func init() {
	serveCmd.AddCommand(proxyCmd)
}
