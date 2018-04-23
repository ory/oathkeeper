/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package cmd

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/meatballhat/negroni-logrus"
	"github.com/ory/graceful"
	"github.com/ory/keto/sdk/go/keto"
	"github.com/ory/metrics-middleware"
	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rsakey"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/sdk/go/oathkeeper"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"
)

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

- HYDRA_CLIENT_ID: The OAuth 2.0 Client ID to be used to connect to ORY Hydra. The client must allowed to request the
	hydra.warden OAuth 2.0 Scope and allowed to access the warden resources.

- HYDRA_CLIENT_SECRET: The OAuth 2.0 Client Secret of the Client ID referenced aboce.

- HYDRA_URL: The URL of ORY Hydra.
	Example: HYDRA_URL=https://hydra.com/

- OATHKEEPER_API_URL: The URL of the Oathkeeper REST API
	Example: OATHKEEPER_API_URL=https://api.oathkeeper.mydomain.com/

- JWT_SHARED_SECRET: The shared secret to be used to encrypt the Authorization Bearer JSON Web Token. Use this
	secret to validate that the Bearer Token was indeed issued by this ORY Oathkeeper instance.

- ISSUER_URL: The public URL where this proxy is listening on.
	Example: ISSUER_URL=https://my-api.com


HTTP(S) CONTROLS
==============

- HTTP_TLS_KEY: Base64 encoded (without padding) string of the private key (PEM encoded) to be used for HTTP over TLS (HTTPS).
	If not set, HTTPS will be disabled and instead HTTP will be served.

- HTTP_TLS_CERT: Base64 encoded (without padding) string of the TLS certificate (PEM encoded) to be used for HTTP over TLS (HTTPS).
	If not set, HTTPS will be disabled and instead HTTP will be served.

- HOST: The host to listen on.
	Default: HOST="" (all interfaces)

- PORT: The port to listen on.
	Default: PORT="4455"


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
		os := oathkeeper.NewSDK(viper.GetString("OATHKEEPER_API_URL"))
		sdk := getHydraSDK()

		issuer := viper.GetString("ISSUER_URL")
		if issuer == "" {
			logger.Fatalln("Please set the issuer URL using the environment variable ISSUER_URL")
		}

		matcher := rule.NewHTTPMatcher(os)
		if err := matcher.Refresh(); err != nil {
			logger.WithError(err).Fatalln("Unable to refresh rules")
		}

		keyManager := &rsakey.HydraManager{
			SDK: sdk,
			Set: viper.GetString("HYDRA_JWK_SET_ID"),
		}

		go refreshRules(matcher, 0)
		go refreshKeys(keyManager, 0)

		eval := proxy.NewJudge(logger, matcher, issuer, newJury(logger))
		d := proxy.NewProxy(eval, logger, keyManager)
		handler := &httputil.ReverseProxy{
			Director:  d.Director,
			Transport: d,
		}

		segmentMiddleware := metrics.NewMetricsManager(
			metrics.Hash(viper.GetString("DATABASE_URL")),
			viper.GetString("DATABASE_URL") != "memory",
			"MSx9A6YQ1qodnkzEFOv22cxOmOCJXMFa",
			[]string{"/"},
			logger,
		)
		go segmentMiddleware.RegisterSegment(Version, GitHash, BuildTime)
		go segmentMiddleware.CommitMemoryStatistics()

		n := negroni.New()
		n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oathkeeper-proxy"))
		n.Use(segmentMiddleware)
		n.UseHandler(handler)

		ch := cors.New(parseCorsOptions("")).Handler(n)

		var cert tls.Certificate
		tlsCert := viper.GetString("HTTP_TLS_CERT")
		tlsKey := viper.GetString("HTTP_TLS_KEY")
		if tlsCert != "" && tlsKey != "" {
			if tlsCert, err := base64.StdEncoding.DecodeString(tlsCert); err != nil {
				logger.WithError(err).Fatalln("Unable to base64 decode the TLS Certificate")
			} else if tlsKey, err := base64.StdEncoding.DecodeString(tlsKey); err != nil {
				logger.WithError(err).Fatalln("Unable to base64 decode the TLS Private Key")
			} else if cert, err = tls.X509KeyPair(tlsCert, tlsKey); err != nil {
				logger.WithError(err).Fatalln("Unable to load X509 key pair")
			}
		}

		addr := fmt.Sprintf("%s:%s", viper.GetString("HOST"), viper.GetString("PORT"))
		server := graceful.WithDefaults(&http.Server{
			Addr:    addr,
			Handler: ch,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		})

		logger.Printf("Listening on %s.\n", addr)
		if err := graceful.Graceful(func() error {
			if tlsCert != "" && tlsKey != "" {
				return server.ListenAndServeTLS("", "")
			}
			return server.ListenAndServe()
		}, server.Shutdown); err != nil {
			logger.Fatalf("Unable to gracefully shutdown HTTP(s) server because %s.\n", err)
			return
		}
		logger.Println("HTTP(s) server was shutdown gracefully")
	},
}

func init() {
	serveCmd.AddCommand(proxyCmd)
}

func newJury(logger logrus.FieldLogger) proxy.Jury {
	k, err := keto.NewCodeGenSDK(&keto.Configuration{
		EndpointURL: viper.GetString("KETO_URL"),
	})
	if err != nil {
		logger.WithError(err).Fatalf("Unable to initialize Keto SDK")
		return nil
	}

	h := getHydraSDK()
	return []proxy.Juror{
		proxy.NewJurorPassThrough(logger),
		proxy.NewJurorWardenOAuth2(logger, k, false, viper.GetString("ANONYMOUS_SUBJECT_ID")),
		proxy.NewJurorWardenOAuth2(logger, k, true, viper.GetString("ANONYMOUS_SUBJECT_ID")),
		proxy.NewJurorOAuth2(logger, h, false, viper.GetString("ANONYMOUS_SUBJECT_ID")),
		proxy.NewJurorOAuth2(logger, h, true, viper.GetString("ANONYMOUS_SUBJECT_ID")),
	}
}
