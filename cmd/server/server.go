package server

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/urfave/negroni"

	"github.com/ory/analytics-go/v4"
	"github.com/ory/graceful"
	"github.com/ory/viper"

	"github.com/ory/x/healthx"
	"github.com/ory/x/logrusx"
	telemetry "github.com/ory/x/metricsx"
	"github.com/ory/x/reqlog"
	"github.com/ory/x/tlsx"

	"github.com/ory/oathkeeper/api"
	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/metrics"
	"github.com/ory/oathkeeper/x"
)

func runProxy(d driver.Driver, n *negroni.Negroni, logger *logrusx.Logger, prom *metrics.PrometheusRepository) func() error {
	return func() error {
		proxy := d.Registry().Proxy()

		handler := &httputil.ReverseProxy{
			Director:  proxy.Director,
			Transport: proxy,
		}

		promCollapsePaths := d.Configuration().PrometheusCollapseRequestPaths()

		n.Use(metrics.NewMiddleware(prom, "oathkeeper-proxy").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath).CollapsePaths(promCollapsePaths))
		n.Use(reqlog.NewMiddlewareFromLogger(logger, "oathkeeper-proxy").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath))
		n.UseHandler(handler)

		var h http.Handler
		if ops, b := d.Configuration().CORS("serve.proxy"); b {
			h = cors.New(ops).Handler(n)
		} else {
			h = n
		}
		certs := cert("proxy", logger)

		addr := d.Configuration().ProxyServeAddress()
		server := graceful.WithDefaults(&http.Server{
			Addr:         addr,
			Handler:      h,
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
			logger.Errorf("Unable to gracefully shutdown HTTP(s) server because %v", err)
			return err
		}
		logger.Println("HTTP(s) server was shutdown gracefully")
		return nil
	}
}

func runAPI(d driver.Driver, n *negroni.Negroni, logger *logrusx.Logger, prom *metrics.PrometheusRepository) func() error {
	return func() error {
		router := x.NewAPIRouter()
		d.Registry().RuleHandler().SetRoutes(router)
		d.Registry().HealthHandler().SetHealthRoutes(router.Router, true)
		d.Registry().CredentialHandler().SetRoutes(router)

		promCollapsePaths := d.Configuration().PrometheusCollapseRequestPaths()

		n.Use(metrics.NewMiddleware(prom, "oathkeeper-api").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath).CollapsePaths(promCollapsePaths))
		n.Use(reqlog.NewMiddlewareFromLogger(logger, "oathkeeper-api").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath))
		n.Use(d.Registry().DecisionHandler()) // This needs to be the last entry, otherwise the judge API won't work

		n.UseHandler(router)

		var h http.Handler
		if ops, b := d.Configuration().CORS("serve.api"); b {
			h = cors.New(ops).Handler(n)
		} else {
			h = n
		}
		certs := cert("api", logger)
		addr := d.Configuration().APIServeAddress()
		server := graceful.WithDefaults(&http.Server{
			Addr:         addr,
			Handler:      h,
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
			logger.Errorf("Unable to gracefully shutdown HTTP(s) server because %v", err)
			return err
		}
		logger.Println("HTTP server was shutdown gracefully")
		return nil
	}
}

func runPrometheus(d driver.Driver, logger *logrusx.Logger, prom *metrics.PrometheusRepository) func() error {
	return func() error {
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
			logger.Errorf("Unable to gracefully shutdown HTTP(s) server because %v", err)
			return err
		}
		logger.Println("HTTP server was shutdown gracefully")
		return nil
	}
}

func cert(daemon string, logger *logrusx.Logger) []tls.Certificate {
	cert, err := tlsx.Certificate(
		viper.GetString("serve."+daemon+".tls.cert.base64"),
		viper.GetString("serve."+daemon+".tls.key.base64"),
		viper.GetString("serve."+daemon+".tls.cert.path"),
		viper.GetString("serve."+daemon+".tls.key.path"),
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
	if err := json.NewEncoder(&id).Encode(viper.AllSettings()); err != nil {
		for _, repo := range c.AccessRuleRepositories() {
			_, _ = id.WriteString(repo.String())
		}
		_, _ = id.WriteString(c.ProxyServeAddress())
		_, _ = id.WriteString(c.APIServeAddress())
		_, _ = id.WriteString(viper.GetString("mutators.id_token.config.jwks_url"))
		_, _ = id.WriteString(viper.GetString("mutators.id_token.config.issuer_url"))
		_, _ = id.WriteString(viper.GetString("authenticators.jwt.config.jwks_urls"))
	}

	return id.String()
}

func isDevelopment(c configuration.Provider) bool {
	return len(c.AccessRuleRepositories()) == 0
}

func RunServe(version, build, date string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		fmt.Println(banner(version))

		logger := logrusx.New("ORY Oathkeeper", version)
		d, err := driver.NewDefaultDriver(cmd.Context(), logger, version, build, date)
		if err != nil {
			return err
		}
		d.Registry().Init()

		adminmw := negroni.New()
		publicmw := negroni.New()

		telemetry := telemetry.New(cmd, logger, d.Configuration().Source(),
			&telemetry.Options{
				Service:       "ory-oathkeeper",
				ClusterID:     telemetry.Hash(clusterID(d.Configuration())),
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
					BatchSize:            250,
					Interval:             time.Hour * 24,
				},
			},
		)

		adminmw.Use(telemetry)
		publicmw.Use(telemetry)

		if tracer := d.Registry().Tracer(); tracer.IsLoaded() {
			adminmw.Use(tracer)
			publicmw.Use(tracer)
		}

		prometheusRepo := metrics.NewPrometheusRepository(logger)

		var eg errgroup.Group
		eg.Go(runAPI(d, adminmw, logger, prometheusRepo))
		eg.Go(runProxy(d, publicmw, logger, prometheusRepo))
		eg.Go(runPrometheus(d, logger, prometheusRepo))

		return eg.Wait()
	}
}
