package cmd

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/meatballhat/negroni-logrus"
	"github.com/ory/graceful"
	"github.com/ory/oathkeeper/evaluator"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/telemetry"
	"github.com/pborman/uuid"
	"github.com/rs/cors"
	"github.com/segmentio/analytics-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
	"net/http"

	"github.com/ory/oathkeeper/extauth"
)

// extauthCmd represents the management command
var extauthCmd = &cobra.Command{
	Use:   "extauth",
	Short: "Starts the ORY Oathkeeper extauth REST API",
	Long: `This starts a HTTP/2 REST API to check if a token is valid and if the token subject is allowed to perform an action on a resource.

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

- ISSUER_URL: The public URL where this proxy is listening on.
	Example: ISSUER_URL=https://my-api.com

HTTP(S) CONTROLS
==============

- EXTAUTH_HOST: The host to listen on.
	Default: EXTAUTH_HOST="" (all interfaces)

- EXTAUTH_PORT: The port to listen on.
	Default: EXTAUTH_PORT="4457"


OTHER CONTROLS
==============
- RULES_REFRESH_INTERVAL: ORY Oathkeeper stores rules in memory for faster access. This value sets the database refresh interval.
	Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Default: RULES_REFRESH_INTERVAL=5s

- JWK_REFRESH_INTERVAL: ORY Oathkeeper stores JSON Web Keys for ID Token signing in memory. This value sets the refresh interval.
	Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Default: JWK_REFRESH_INTERVAL=5m

- HYDRA_JWK_SET_ID: The JSON Web Key set identifier that will be used to create, store, and retrieve the JSON Web Key from ORY Hydra.
	Default: HYDRA_JWK_SET_ID=oathkeeper:id-token
` + corsMessage,
	Run: func(cmd *cobra.Command, args []string) {
		rules, err := newRuleManager(viper.GetString("DATABASE_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to rule backend")
		}

		config := &proxyConfig{rules: rules}
		runExtauth(config)
	},
}

func runExtauth(c *proxyConfig) {
	sdk := getHydraSDK()

	issuer := viper.GetString("ISSUER_URL")
	if issuer == "" {
		logger.Fatalln("Please set the issuer URL using the environment variable ISSUER_URL")
	}

	matcher := &rule.CachedMatcher{Manager: c.rules, Rules: []rule.Rule{}}

	if err := matcher.Refresh(); err != nil {
		logger.WithError(err).Fatalln("Unable to refresh rules")
	}

	segmentMiddleware := new(telemetry.Middleware)
	segment := telemetry.Manager{
		Segment:      analytics.New("MSx9A6YQ1qodnkzEFOv22cxOmOCJXMFa"),
		Middleware:   segmentMiddleware,
		ID:           issuer,
		BuildVersion: Version,
		BuildTime:    BuildTime,
		BuildHash:    GitHash,
		Logger:       logger,
		InstanceID:   uuid.New(),
	}

	go segment.Identify()
	go segment.Submit()
	go refreshRules(c, matcher, 0)

	eval := evaluator.NewWardenEvaluator(logger, matcher, sdk, issuer)
	handler := extauth.Handler{Evaluator: eval}
	router := httprouter.New()
	handler.SetRoutes(router)

	n := negroni.New()
	n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oathkeeper-extauth"))
	n.UseHandler(router)

	ch := cors.New(parseCorsOptions(c.corsPrefix)).Handler(n)

	addr := fmt.Sprintf("%s:%s", viper.GetString("EXTAUTH_HOST"), viper.GetString("EXTAUTH_PORT"))
	server := graceful.WithDefaults(&http.Server{
		Addr:    addr,
		Handler: ch,
	})

	logger.Printf("Listening on %s.\n", addr)
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		logger.Fatalf("Unable to gracefully shutdown HTTP server because %s.\n", err)
		return
	}
	logger.Println("HTTP server was shutdown gracefully")
}

func init() {
	serveCmd.AddCommand(extauthCmd)
}
