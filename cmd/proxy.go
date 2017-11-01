package cmd

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

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

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Runs the reverse proxy",
	Run: func(cmd *cobra.Command, args []string) {
		sdk, err := hydra.NewSDK(&hydra.Configuration{
			ClientID:     viper.GetString("HYDRA_CLIENT_ID"),
			ClientSecret: viper.GetString("HYDRA_CLIENT_SECRET"),
			EndpointURL:  viper.GetString("HYDRA_URL"),
			Scopes:       []string{"hydra.warden"},
		})
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to Hydra SDK.")
			return
		}
		backend, err := url.Parse(viper.GetString("BACKEND_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to parse backend URL.")
		}

		rm, err := newRuleManager(viper.GetString("DATABASE_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to rule backend.")
		}

		matcher := &rule.CachedMatcher{Manager: rm, Rules: []rule.Rule{}}

		if err := matcher.Refresh(); err != nil {
			logger.WithError(err).Fatalln("Unable to refresh rules.")
		}

		go refresh(matcher, 0)

		eval := evaluator.NewWardenEvaluator(logger, matcher, sdk)
		d := director.NewDirector(backend, eval, logger, viper.GetString("JWT_SHARED_SECRET"))
		proxy := &httputil.ReverseProxy{
			Director:  d.Director,
			Transport: d,
		}

		n := negroni.New()
		n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oahtkeeper-proxy"))
		n.UseHandler(proxy)

		allowCredentials, _ := strconv.ParseBool(os.Getenv("CORS_ALLOWED_CREDENTIALS"))
		debug, _ := strconv.ParseBool(os.Getenv("CORS_DEBUG"))
		maxAge, _ := strconv.Atoi(os.Getenv("CORS_MAX_AGE"))
		ch := cors.New(cors.Options{
			AllowedOrigins:   strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ","),
			AllowedMethods:   strings.Split(os.Getenv("CORS_ALLOWED_METHODS"), ","),
			AllowedHeaders:   strings.Split(os.Getenv("CORS_ALLOWED_HEADERS"), ","),
			ExposedHeaders:   strings.Split(os.Getenv("CORS_EXPOSED_HEADERS"), ","),
			AllowCredentials: allowCredentials,
			MaxAge:           maxAge,
			Debug:            debug,
		}).Handler(n)

		addr := fmt.Sprintf("%s:%s", viper.GetString("PROXY_HOST"), viper.GetString("PROXY_PORT"))
		server := graceful.WithDefaults(&http.Server{
			Addr:    addr,
			Handler: ch,
		})

		logger.Printf("Listening on %s.\n", addr)
		if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
			logger.Fatalf("Unable to gracefully shutdown HTTP server becase %s.\n", err)
			return
		}
		logger.Println("HTTP server was shutdown gracefully")
	},
}

func refresh(m *rule.CachedMatcher, fails int) {
	duration, _ := time.ParseDuration(viper.GetString("REFRESH_DELAY"))
	if duration == 0 {
		duration = time.Second * 30
	}

	time.Sleep(duration)

	if err := m.Refresh(); err != nil {
		logger.WithError(err).WithField("retry", fails).Errorln("Unable to refresh rules.")
		if fails > 15 {
			logger.WithError(err).WithField("retry", fails).Fatalf("Terminating after retry %d.\n", fails)
		}

		refresh(m, fails+1)
		return
	}

	refresh(m, 0)
}

func init() {
	RootCmd.AddCommand(proxyCmd)
}
