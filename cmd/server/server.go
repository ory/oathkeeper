package server

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"

	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/urfave/negroni"

	"github.com/ory/viper"

	"github.com/ory/x/healthx"

	"github.com/ory/graceful"
	"github.com/ory/x/corsx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/metricsx"
	"github.com/ory/x/tlsx"

	"github.com/ory/oathkeeper/api"
	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/x"
)

func runProxy(d driver.Driver, n *negroni.Negroni, logger *logrus.Logger) func() {
	return func() {
		proxy := d.Registry().Proxy()

		handler := &httputil.ReverseProxy{
			Director:  proxy.Director,
			Transport: proxy,
		}

		n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oathkeeper-proxy"))
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

func runAPI(d driver.Driver, n *negroni.Negroni, logger *logrus.Logger) func() {
	return func() {
		router := x.NewAPIRouter()
		d.Registry().RuleHandler().SetRoutes(router)
		d.Registry().HealthHandler().SetRoutes(router.Router, true)
		d.Registry().CredentialHandler().SetRoutes(router)

		n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oathkeeper-api"))
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

		metrics := metricsx.New(cmd, logger,
			&metricsx.Options{
				Service:       "ory-oathkeeper",
				ClusterID:     clusterID(d.Configuration()),
				IsDevelopment: isDevelopment(d.Configuration()),
				WriteKey:      "MSx9A6YQ1qodnkzEFOv22cxOmOCJXMFa",
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
			},
		)

		adminmw.Use(metrics)
		publicmw.Use(metrics)

		var wg sync.WaitGroup
		tasks := []func(){
			runAPI(d, adminmw, logger),
			runProxy(d, publicmw, logger),
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
