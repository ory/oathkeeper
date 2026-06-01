// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/urfave/negroni"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/ory/analytics-go/v5"
	"github.com/ory/graceful"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/corsx"
	"github.com/ory/x/healthx"
	"github.com/ory/x/httprouterx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/metricsx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/reqlog"
	"github.com/ory/x/serverx"
	"github.com/ory/x/tlsx"
	"github.com/ory/x/urlx"

	"github.com/ory/oathkeeper/api"
	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/metrics"
	"github.com/ory/oathkeeper/x"
)

func isTimeoutError(err error) bool {
	var te interface{ Timeout() bool } = nil
	return errors.As(err, &te) && te.Timeout() || errors.Is(err, context.DeadlineExceeded)
}

func runProxy(d driver.Registry, n *negroni.Negroni, logger *logrusx.Logger, prom *metrics.PrometheusRepository) func() {
	return func() {
		proxy := d.Proxy()
		transport := otelhttp.NewTransport(proxy, otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string { return "upstream" }))
		proxyHandler := &httputil.ReverseProxy{
			Rewrite:   proxy.Rewrite,
			Transport: transport,
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				switch {
				case errors.Is(r.Context().Err(), context.Canceled):
					logger.WithError(err).Warn("http: client canceled request")
					w.WriteHeader(499) // http://nginx.org/en/docs/dev/development_guide.html
				case isTimeoutError(err):
					logger.WithError(err).Errorf("http: gateway timeout")
					w.WriteHeader(http.StatusGatewayTimeout)
				default:
					logger.WithError(err).Errorf("http: gateway error")
					w.WriteHeader(http.StatusBadGateway)
				}
			},
		}

		promHidePaths := d.Config().PrometheusHideRequestPaths()
		promCollapsePaths := d.Config().PrometheusCollapseRequestPaths()
		n.Use(metrics.NewMiddleware(prom, "oathkeeper-proxy").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath).HidePaths(promHidePaths).CollapsePaths(promCollapsePaths))
		n.Use(reqlog.NewMiddlewareFromLogger(logger, "oathkeeper-proxy").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath))
		n.Use(corsx.ContextualizedMiddleware(func(ctx context.Context) (opts cors.Options, enabled bool) { //nolint:staticcheck // legacy middleware still supported
			return d.Config().CORS("proxy")
		}))

		n.UseHandler(proxyHandler)

		certs := cert(d.Config(), "proxy", logger)

		addr := d.Config().ProxyServeAddress()
		server := graceful.WithDefaults(&http.Server{ //nolint:gosec // server intentionally configured by graceful defaults
			Addr:         addr,
			Handler:      otelx.NewMiddleware(n, "proxy"),
			TLSConfig:    &tls.Config{Certificates: certs}, //nolint:gosec // TLS settings handled via configuration
			ReadTimeout:  d.Config().ProxyReadTimeout(),
			WriteTimeout: d.Config().ProxyWriteTimeout(),
			IdleTimeout:  d.Config().ProxyIdleTimeout(),
		})

		if err := graceful.Graceful(func() error {
			if certs != nil {
				logger.Printf("Listening on https://%s", addr)
				return server.ListenAndServeTLS("", "")
			}
			logger.Infof("Listening on http://%s", addr)
			return server.ListenAndServe()
		}, server.Shutdown); err != nil {
			logger.Fatalf("Unable to gracefully shutdown HTTP(s) server because %v", err)
			return
		}
		logger.Println("HTTP(s) server was shutdown gracefully")
	}
}

func runAPI(d driver.Registry, n *negroni.Negroni, logger *logrusx.Logger, prom *metrics.PrometheusRepository) func() {
	return func() {
		router := httprouterx.NewRouter()
		d.RuleHandler().SetRoutes(router)
		d.HealthHandler().SetHealthRoutes(router, true)
		d.CredentialHandler().SetRoutes(router)
		d.DecisionHandler().SetRoutes(router)
		router.Handle("/", serverx.DefaultNotFoundHandler)

		promHidePaths := d.Config().PrometheusHideRequestPaths()
		promCollapsePaths := d.Config().PrometheusCollapseRequestPaths()

		n.Use(metrics.NewMiddleware(prom, "oathkeeper-api").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath).HidePaths(promHidePaths).CollapsePaths(promCollapsePaths))
		n.Use(reqlog.NewMiddlewareFromLogger(logger, "oathkeeper-api").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath))
		n.Use(corsx.ContextualizedMiddleware(func(ctx context.Context) (opts cors.Options, enabled bool) { //nolint:staticcheck // legacy middleware still supported
			return d.Config().CORS("api")
		}))

		n.UseHandler(router)

		certs := cert(d.Config(), "api", logger)
		addr := d.Config().APIServeAddress()
		server := graceful.WithDefaults(&http.Server{ //nolint:gosec // server intentionally configured by graceful defaults
			Addr:         addr,
			Handler:      otelx.NewMiddleware(n, "api"),
			TLSConfig:    &tls.Config{Certificates: certs}, //nolint:gosec // TLS settings handled via configuration
			ReadTimeout:  d.Config().APIReadTimeout(),
			WriteTimeout: d.Config().APIWriteTimeout(),
			IdleTimeout:  d.Config().APIIdleTimeout(),
		})

		if err := graceful.Graceful(func() error {
			if certs != nil {
				logger.Printf("Listening on https://%s", addr)
				return server.ListenAndServeTLS("", "")
			}
			logger.Infof("Listening on http://%s", addr)
			return server.ListenAndServe()
		}, server.Shutdown); err != nil {
			logger.Errorf("unable to gracefully shutdown HTTP(s) server: %v", err)
			return
		}
		logger.Println("HTTP server was shutdown gracefully")
	}
}

func runPrometheus(d driver.Registry, logger *logrusx.Logger, prom *metrics.PrometheusRepository) func() {
	return func() {
		promPath := d.Config().PrometheusMetricsPath()
		promAddr := d.Config().PrometheusServeAddress()

		server := graceful.WithDefaults(&http.Server{ //nolint:gosec // server intentionally configured by graceful defaults
			Addr:    promAddr,
			Handler: promhttp.HandlerFor(prom.Registry, promhttp.HandlerOpts{}),
		})

		http.Handle(promPath, promhttp.Handler())
		// Expose the registered metrics via HTTP.
		if err := graceful.Graceful(func() error {
			logger.Infof("Listening on http://%s", promAddr)
			return server.ListenAndServe()
		}, server.Shutdown); err != nil {
			logger.Errorf("Unable to gracefully shutdown HTTP(s) server: %v", err)
			return
		}
		logger.Println("HTTP server was shutdown gracefully")
	}
}

func cert(config configuration.Configuration, daemon string, logger *logrusx.Logger) []tls.Certificate {
	tlsCfg := config.TLSConfig(daemon)

	cert, err := tlsx.Certificate(
		tlsCfg.Cert.Base64,
		tlsCfg.Key.Base64,
		tlsCfg.Cert.Path,
		tlsCfg.Key.Path,
	)

	if err == nil {
		logger.Infof("Setting up HTTPS for %s", daemon)
		return cert
	} else if errors.Cause(err) != tlsx.ErrNoCertificatesConfigured {
		logger.WithError(err).Fatalf("Unable to load HTTPS TLS Certificate")
	}

	logger.Infof("TLS has not been configured for %s, skipping", daemon)
	return nil
}

func clusterID(c configuration.Configuration) string {
	var id bytes.Buffer
	if err := json.NewEncoder(&id).Encode(c.AllSettings()); err != nil {
		for _, repo := range c.AccessRuleRepositories() {
			_, _ = id.WriteString(repo.String())
		}
		_, _ = id.WriteString(c.ProxyServeAddress())
		_, _ = id.WriteString(c.APIServeAddress())
		_, _ = id.WriteString(c.String("mutators.id_token.config.jwks_url"))
		_, _ = id.WriteString(c.String("mutators.id_token.config.issuer_url"))
		_, _ = id.WriteString(c.String("authenticators.jwt.config.jwks_urls"))
	}

	return id.String()
}

func isDevelopment(c configuration.Configuration) bool {
	return len(c.AccessRuleRepositories()) == 0
}

func RunServe(cmd *cobra.Command, _ []string) error {
	fmt.Println(banner(x.Version))

	logger := logrusx.New("Ory Oathkeeper", x.Version)
	reg, err := driver.NewDefaultDriver(cmd.Context(), logger, cmd.Flags())
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "unable to initialize registry: %v\n", err)
		return cmdx.FailSilently(cmd)
	}
	reg.Init()

	urls := []string{
		reg.Config().APIServeAddress(),
		reg.Config().ProxyServeAddress(),
	}

	if c, y := reg.Config().CORS("api"); y {
		urls = append(urls, c.AllowedOrigins...)
	}
	if c, y := reg.Config().CORS("proxy"); y {
		urls = append(urls, c.AllowedOrigins...)
	}

	host := urlx.ExtractPublicAddress(urls...)

	telemetry := metricsx.New(cmd, logger, reg.Config().Source(), &metricsx.Options{
		Service:       "oathkeeper",
		DeploymentId:  metricsx.Hash(clusterID(reg.Config())),
		IsDevelopment: isDevelopment(reg.Config()),
		WriteKey:      "xRVRP48SAKw6ViJEnvB0u2PY8bVlsO6O",
		WhitelistedPaths: []string{
			"/",
			api.CredentialsPath,
			api.DecisionPath,
			api.RulesPath,
			healthx.VersionPath,
			healthx.AliveCheckPath,
			healthx.ReadyCheckPath,
		},
		BuildVersion: x.Version,
		BuildTime:    x.Date,
		BuildHash:    x.Commit,
		Config: &analytics.Config{
			Endpoint:             "https://sqa.ory.sh",
			GzipCompressionLevel: 6,
			BatchMaxSize:         500 * 1000,
			BatchSize:            1000,
			Interval:             time.Hour * 6,
		},
		Hostname: host,
	})

	recovery := negroni.NewRecovery()
	recovery.Logger = logger

	adminmw := negroni.New(
		recovery,
		telemetry,
	)
	publicmw := negroni.New(
		recovery,
		telemetry,
	)

	prometheusRepo := metrics.NewConfigurablePrometheusRepository(reg)
	var wg sync.WaitGroup
	tasks := []func(){
		runAPI(reg, adminmw, logger, prometheusRepo),
		runProxy(reg, publicmw, logger, prometheusRepo),
		runPrometheus(reg, logger, prometheusRepo),
	}
	wg.Add(len(tasks))
	for _, t := range tasks {
		go func(t func()) {
			defer wg.Done()
			t()
		}(t)
	}
	wg.Wait()
	return nil
}
