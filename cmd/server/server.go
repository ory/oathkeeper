package server

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/urfave/negroni"

	"github.com/ory/analytics-go/v4"
	"github.com/ory/graceful"
	"github.com/ory/viper"

	"github.com/ory/x/corsx"
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

func runProxy(d driver.Driver, n *negroni.Negroni, logger *logrus.Logger, prom *metrics.PrometheusRepository) func() {
	return func() {
		proxy := d.Registry().Proxy()

		handler := &httputil.ReverseProxy{
			Director:  proxy.Director,
			Transport: proxy,
		}

		n.Use(metrics.NewMiddleware(prom, "oathkeeper-proxy").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath))
		n.Use(reqlog.NewMiddlewareFromLogger(logger, "oathkeeper-proxy").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath))
		n.UseHandler(handler)

		h := corsx.Initialize(n, logger, "serve.proxy")
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
			logger.Fatalf("Unable to gracefully shutdown HTTP(s) server because %v", err)
			return
		}
		logger.Println("HTTP(s) server was shutdown gracefully")
	}
}

func runAPI(d driver.Driver, n *negroni.Negroni, logger *logrus.Logger, prom *metrics.PrometheusRepository) func() {
	return func() {
		router := x.NewAPIRouter()
		d.Registry().RuleHandler().SetRoutes(router)
		d.Registry().HealthHandler().SetRoutes(router.Router, true)
		d.Registry().CredentialHandler().SetRoutes(router)

		n.Use(metrics.NewMiddleware(prom, "oathkeeper-api").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath))
		n.Use(reqlog.NewMiddlewareFromLogger(logger, "oathkeeper-api").ExcludePaths(healthx.ReadyCheckPath, healthx.AliveCheckPath))
		n.Use(d.Registry().DecisionHandler()) // This needs to be the last entry, otherwise the judge API won't work

		n.UseHandler(router)

		h := corsx.Initialize(n, logger, "serve.api")
		certs := cert("api", logger)
		addr := d.Configuration().APIServeAddress()
		server := graceful.WithDefaults(&http.Server{
			Addr:      addr,
			Handler:   h,
			TLSConfig: &tls.Config{Certificates: certs},
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

func runPrometheus(d driver.Driver, logger *logrus.Logger, prom *metrics.PrometheusRepository) func() {
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

func cert(daemon string, logger logrus.FieldLogger) []tls.Certificate {
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

func RunServe(version, build, date string) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		fmt.Println(banner(version))

		logger := logrusx.New()
		d := driver.NewDefaultDriver(logger, version, build, date, true)
		d.Registry().Init()

		adminmw := negroni.New()
		publicmw := negroni.New()

		telemetry := telemetry.New(cmd, logger,
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
