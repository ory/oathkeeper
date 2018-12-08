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
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/meatballhat/negroni-logrus"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"

	"github.com/ory/graceful"
	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/judge"
	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rsakey"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/x/corsx"
	"github.com/ory/x/metricsx"
)

// serveApiCmd represents the management command
var serveApiCmd = &cobra.Command{
	Use:   "api",
	Short: "Starts the ORY Oathkeeper HTTP API",
	Long: `This starts a HTTP/2 REST API for managing ORY Oathkeeper.

CORE CONTROLS
=============

` + databaseUrl + `


` + credentialsIssuer + `


HTTP CONTROLS
==============
` + tlsMessage + `

- HOST: The host to listen on.
	Default: HOST="" (all interfaces)
- PORT: The port to listen on.
	Default: PORT="4456"

` + corsMessage,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := connectToDatabase(viper.GetString("DATABASE_URL"))
		if err != nil {
			logger.WithError(err).Fatalln("Unable to initialize database connectivity")
		}

		rules, err := newRuleManager(db)
		if err != nil {
			logger.WithError(err).Fatalln("Unable to connect to rule backend")
		}

		keyManager, err := keyManagerFactory(logger)
		if err != nil {
			logger.WithError(err).Fatalln("Unable to initialize the ID Token signing algorithm")
		}

		matcher := rule.NewCachedMatcher(rules)

		enabledAuthenticators, enabledAuthorizers, enabledCredentialIssuers := enabledHandlerNames()
		availableAuthenticators, availableAuthorizers, availableCredentialIssuers := availableHandlerNames()

		authenticators, authorizers, credentialIssuers := handlerFactories(keyManager)
		eval := proxy.NewRequestHandler(logger, authenticators, authorizers, credentialIssuers)

		router := httprouter.New()
		writer := herodot.NewJSONWriter(logger)
		ruleHandler := rule.NewHandler(writer, rules, rule.ValidateRule(
			enabledAuthenticators, availableAuthenticators,
			enabledAuthorizers, availableAuthorizers,
			enabledCredentialIssuers, availableCredentialIssuers,
		))
		judgeHandler := judge.NewHandler(eval, logger, matcher, router)
		keyHandler := rsakey.NewHandler(writer, keyManager)
		health := newHealthHandler(db, writer, router)
		ruleHandler.SetRoutes(router)
		keyHandler.SetRoutes(router)
		health.SetRoutes(router)

		n := negroni.New()
		n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oathkeeper-api"))

		if ok, _ := cmd.Flags().GetBool("disable-telemetry"); !ok {
			logger.Println("Transmission of telemetry data is enabled, to learn more go to: https://www.ory.sh/docs/ecosystem/sqa")

			segmentMiddleware := metricsx.NewMetricsManager(
				metricsx.Hash(viper.GetString("DATABASE_URL")),
				viper.GetString("DATABASE_URL") != "memory",
				"MSx9A6YQ1qodnkzEFOv22cxOmOCJXMFa",
				[]string{"/rules", "/.well-known/jwks.json"},
				logger,
				"ory-oathkeeper-api",
				100,
				"",
			)
			go segmentMiddleware.RegisterSegment(Version, GitHash, BuildTime)
			go segmentMiddleware.CommitMemoryStatistics()
			n.Use(segmentMiddleware)
		}

		n.UseHandler(judgeHandler)
		var h http.Handler = n
		if viper.GetString("CORS_ENABLED") == "true" {
			h = cors.New(corsx.ParseOptions()).Handler(n)
		}

		go refreshKeys(keyManager, 0)
		go refreshRules(matcher, 0)

		cert, err := getTLSCertAndKey()
		if err != nil {
			logger.Fatalf("%v", err)
		}

		certs := []tls.Certificate{}
		if cert != nil {
			certs = append(certs, *cert)
		}

		addr := fmt.Sprintf("%s:%s", viper.GetString("HOST"), viper.GetString("PORT"))
		server := graceful.WithDefaults(&http.Server{
			Addr:    addr,
			Handler: h,
			TLSConfig: &tls.Config{
				Certificates: certs,
			},
		})

		if err := graceful.Graceful(func() error {
			if cert != nil {
				logger.Printf("Listening on https://%s.\n", addr)
				return server.ListenAndServeTLS("", "")
			}
			logger.Printf("Listening on http://%s.\n", addr)
			return server.ListenAndServe()
		}, server.Shutdown); err != nil {
			logger.Fatalf("Unable to gracefully shutdown HTTP(s) server because %v.\n", err)
			return
		}
		logger.Println("HTTP server was shutdown gracefully")
	},
}

func init() {
	serveCmd.AddCommand(serveApiCmd)
}
