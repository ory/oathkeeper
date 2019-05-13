package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httputil"
	"sync"

	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"

	"github.com/ory/graceful"
	"github.com/ory/oathkeeper/api"
	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/corsx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/metricsx"
	"github.com/ory/x/tlsx"
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
			logger.Printf("Listening on http://%s", addr)
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

		n.With(
			negronilogrus.NewMiddlewareFromLogger(logger, "oathkeeper-api"),
			d.Registry().JudgeHandler(), // This needs to be the last entry, otherwise the judge API won't work
		)
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
			logger.Printf("Listening on http://%s", addr)
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

	logger.Info("TLS has not been configured for %s, skipping", daemon)
	return nil
}

func RunServe(version, build, date string) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		fmt.Println(banner(version))

		logger := logrusx.New()
		d := driver.NewDefaultDriver(logger, version, build, date, true)
		if err := d.Registry().Init(); err != nil {
			logger.WithError(err).Fatal("Unable to initialize.")
		}

		adminmw := negroni.New()
		publicmw := negroni.New()

		metrics := metricsx.New(cmd, logger,
			&metricsx.Options{
				Service:       "ory-oathkeeper",
				ClusterID:     metricsx.Hash(viper.GetString("DATABASE_URL")), // TODO
				IsDevelopment: viper.GetString("DATABASE_URL") != "memory",    // TODO
				WriteKey:      "MSx9A6YQ1qodnkzEFOv22cxOmOCJXMFa",
				WhitelistedPaths: []string{
					"/",
					api.CredentialsPath,
				},
				BuildVersion: version,
				BuildTime:    build,
				BuildHash:    date,
			},
		)

		adminmw.Use(metrics)
		publicmw.Use(metrics)

		var wg sync.WaitGroup
		tasks := []func(){runAPI(d, adminmw, logger), runProxy(d, publicmw, logger)}
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
