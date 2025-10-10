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
	"github.com/ory/x/corsx"
	"github.com/ory/x/healthx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/metricsx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/reqlog"
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

func runProxy(d driver.Driver, n *negroni.Negroni, logger *logrusx.Logger, prom *metrics.PrometheusRepository) func() {
	return func() {
		proxy := d.Registry().Proxy()
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

		promHidePaths := d.Configuration().PrometheusHideRequestPaths()
		promCollapsePaths := d.Configuration().PrometheusCollapseRequestPaths()
		n.Use(metrics.NewMiddleware(prom, "oathkeeper-proxy").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath).HidePaths(promHidePaths).CollapsePaths(promCollapsePaths))
		n.Use(reqlog.NewMiddlewareFromLogger(logger, "oathkeeper-proxy").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath))
		n.Use(corsx.ContextualizedMiddleware(func(ctx context.Context) (opts cors.Options, enabled bool) {
			return d.Configuration().CORS("proxy")
		}))

		n.UseHandler(proxyHandler)

		certs := cert(d.Configuration(), "proxy", logger)

		addr := d.Configuration().ProxyServeAddress()
		server := graceful.WithDefaults(&http.Server{
			Addr:         addr,
			Handler:      otelx.NewHandler(n, "proxy"),
			TLSConfig:    &tls.Config{Certificates: certs},
			ReadTimeout:  d.Configuration().ProxyReadTimeout(),
			WriteTimeout: d.Configuration().ProxyWriteTimeout(),
			IdleTimeout:  d.Configuration().ProxyIdleTimeout(),
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

func runAPI(d driver.Driver, n *negroni.Negroni, logger *logrusx.Logger, prom *metrics.PrometheusRepository) func() {
	return func() {
		router := x.NewAPIRouter()
		d.Registry().RuleHandler().SetRoutes(router)
		d.Registry().HealthHandler().SetHealthRoutes(router.Router, true)
		d.Registry().CredentialHandler().SetRoutes(router)

		promHidePaths := d.Configuration().PrometheusHideRequestPaths()
		promCollapsePaths := d.Configuration().PrometheusCollapseRequestPaths()

		n.Use(metrics.NewMiddleware(prom, "oathkeeper-api").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath).HidePaths(promHidePaths).CollapsePaths(promCollapsePaths))
		n.Use(reqlog.NewMiddlewareFromLogger(logger, "oathkeeper-api").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath))
		n.Use(corsx.ContextualizedMiddleware(func(ctx context.Context) (opts cors.Options, enabled bool) {
			return d.Configuration().CORS("api")
		}))
		n.Use(d.Registry().DecisionHandler()) // This needs to be the last entry, otherwise the judge API won't work

		n.UseHandler(router)

		certs := cert(d.Configuration(), "api", logger)
		addr := d.Configuration().APIServeAddress()
		server := graceful.WithDefaults(&http.Server{
			Addr:         addr,
			Handler:      otelx.TraceHandler(n),
			TLSConfig:    &tls.Config{Certificates: certs},
			ReadTimeout:  d.Configuration().APIReadTimeout(),
			WriteTimeout: d.Configuration().APIWriteTimeout(),
			IdleTimeout:  d.Configuration().APIIdleTimeout(),
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
		logger.Println("HTTP server was shutdown gracefully")
	}
}

func runPrometheus(d driver.Driver, logger *logrusx.Logger, prom *metrics.PrometheusRepository) func() {
	return func() {
		promPath := d.Configuration().PrometheusMetricsPath()
		promAddr := d.Configuration().PrometheusServeAddress()

		server := graceful.WithDefaults(&http.Server{
			Addr:    promAddr,
			Handler: promhttp.HandlerFor(prom.Registry, promhttp.HandlerOpts{}),
		})

		http.Handle(promPath, promhttp.Handler())
		// Expose the registered metrics via HTTP.
		if err := graceful.Graceful(func() error {
			logger.Infof("Listening on http://%s", promAddr)
			return server.ListenAndServe()
		}, server.Shutdown); err != nil {
			logger.Fatalf("Unable to gracefully shutdown HTTP(s) server because %v", err)
			return
		}
		logger.Println("HTTP server was shutdown gracefully")
	}
}

func cert(config configuration.Provider, daemon string, logger *logrusx.Logger) []tls.Certificate {
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

func clusterID(c configuration.Provider) string {
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

func isDevelopment(c configuration.Provider) bool {
	return len(c.AccessRuleRepositories()) == 0
}

func RunServe(version, build, date string) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, _ []string) {
		fmt.Println(banner(version))

		logger := logrusx.New("ORY Oathkeeper", version)
		d := driver.NewDefaultDriver(logger, version, build, date, cmd.Flags())
		d.Registry().Init()

		adminmw := negroni.New()
		publicmw := negroni.New()

		urls := []string{
			d.Configuration().APIServeAddress(),
			d.Configuration().ProxyServeAddress(),
		}

		if c, y := d.Configuration().CORS("api"); y {
			urls = append(urls, c.AllowedOrigins...)
		}
		if c, y := d.Configuration().CORS("proxy"); y {
			urls = append(urls, c.AllowedOrigins...)
		}

		host := urlx.ExtractPublicAddress(urls...)

		telemetry := metricsx.New(cmd, logger, d.Configuration().Source(), &metricsx.Options{
			Service:       "oathkeeper",
			DeploymentId:  metricsx.Hash(clusterID(d.Configuration())),
			IsDevelopment: isDevelopment(d.Configuration()),
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
			BuildVersion: version,
			BuildTime:    build,
			BuildHash:    date,
			Config: &analytics.Config{
				Endpoint:             "https://sqa.ory.sh",
				GzipCompressionLevel: 6,
				BatchMaxSize:         500 * 1000,
				BatchSize:            1000,
				Interval:             time.Hour * 6,
			},
			Hostname: host,
		})

		adminmw.Use(telemetry)
		publicmw.Use(telemetry)

		prometheusRepo := metrics.NewConfigurablePrometheusRepository(d, logger)
		var wg sync.WaitGroup
		tasks := []func(){
			runAPI(d, adminmw, logger, prometheusRepo),
			runProxy(d, publicmw, logger, prometheusRepo),
			runPrometheus(d, logger, prometheusRepo),
		}
		wg.Add(len(tasks))
		for _, t := range tasks {
			go func(t func()) {
				defer wg.Done()
				t()
			}(t)
		}
		wg.Wait()
	}
}
