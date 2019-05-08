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
	"net/http/httputil"
	"net/url"

	"github.com/ory/oathkeeper/pipeline/authz"

	"github.com/ory/x/urlx"

	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/urfave/negroni"

	"github.com/ory/graceful"
	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/x/corsx"
	"github.com/ory/x/metricsx"
)

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Starts the ORY Oathkeeper firewall reverse proxy",
	Long: `This starts a HTTP/2 reverse proxy capable of deciding whether to forward API requests or to block them based on a set of rules.

This command exposes a variety of controls via environment variables. You can
set environments using "export KEY=VALUE" (Linux/macOS) or "set KEY=VALUE" (Windows). On Linux,
you can also set environments by prepending key value pairs: "KEY=VALUE KEY2=VALUE2 oathkeeper"

All possible controls are listed below.

REQUIRED CONTROLS
=============

- OATHKEEPER_API_URL: The URL of the Oathkeeper REST API
	--------------------------------------------------------------
	Example: OATHKEEPER_API_URL=https://api.oathkeeper.mydomain.com/
	--------------------------------------------------------------


HTTP(S) CONTROLS
==============
` + tlsMessage + `

- HOST: The host to listen on.
	--------------------------------------------------------------
	Default: HOST="" (all interfaces)
	--------------------------------------------------------------

- PORT: The port to listen on.
	--------------------------------------------------------------
	Default: PORT="4455"
	--------------------------------------------------------------


AUTHENTICATORS
==============

- JSON Web Token Authenticator:
	- AUTHENTICATOR_JWT_JWKS_URL: The URL where ORY Oathkeeper can retrieve JSON Web Keys from for validating
		the JSON Web Token. Usually something like "https://my-keys.com/.well-known/jwks.json". The response
		of that endpoint must return a JSON Web Key Set (JWKS).
	- AUTHENTICATOR_JWT_SCOPE_STRATEGY: The strategy to be used to validate the scope claim. Strategies "HIERARCHIC", "EXACT",
		"WILDCARD", "NONE" are supported. For more information on each strategy, go to http://www.ory.sh/docs.
		--------------------------------------------------------------
		Default: AUTHENTICATOR_JWT_SCOPE_STRATEGY=EXACT
		--------------------------------------------------------------

- OAuth 2.0 Client Credentials Authenticator:
	- AUTHENTICATOR_OAUTH2_CLIENT_CREDENTIALS_TOKEN_URL: Sets the OAuth 2.0 Token URL that should be used to check if
		the provided credentials are valid or not.
		--------------------------------------------------------------
		Example: AUTHENTICATOR_OAUTH2_CLIENT_CREDENTIALS_TOKEN_URL=http://my-oauth2-server/oauth2/token
		--------------------------------------------------------------

- OAuth 2.0 Token Introspection Authenticator:
	- AUTHENTICATOR_OAUTH2_INTROSPECTION_URL: The OAuth 2.0 Token Introspection URL.
		--------------------------------------------------------------
		Example: AUTHENTICATOR_OAUTH2_INTROSPECTION_URL=http://my-oauth2-server/oauth2/introspect
		--------------------------------------------------------------

	- AUTHENTICATOR_OAUTH2_INTROSPECTION_SCOPE_STRATEGY: The strategy to be used to validate the scope claim.
		Strategies "HIERARCHIC", "EXACT", "WILDCARD", "NONE" are supported. For more information on each strategy, go to http://www.ory.sh/docs.
		--------------------------------------------------------------
		Default: AUTHENTICATOR_OAUTH2_INTROSPECTION_SCOPE_STRATEGY=EXACT
		--------------------------------------------------------------

	If the OAuth 2.0 Token Introspection Endpoint itself is protected with OAuth 2.0, you can provide the access credentials to perform
	an OAuth 2.0 Client Credentials flow before accessing ORY Hydra's APIs.

	These settings are usually not required and an optional! If you don't need this feature, leave them undefined.


		- AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_ID: The OAuth 2.0 Client ID the client that performs the OAuth 2.0
			Token Introspection. The OAuth 2.0 Token Introspection endpoint is typically protected and requires a valid
			OAuth 2.0 Client in order to check if a token is valid or not.
			--------------------------------------------------------------
			Example: AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_ID=my-client-id
			--------------------------------------------------------------

		- AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_SECRET: The OAuth 2.0 Client Secret of the client that performs the OAuth 2.0 Token Introspection.
			--------------------------------------------------------------
			Example: AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_ID=my-client-secret
			--------------------------------------------------------------

		- AUTHENTICATOR_OAUTH2_INTROSPECTION_TOKEN_URL: The OAuth 2.0 Token URL.
			--------------------------------------------------------------
			Example: AUTHENTICATOR_OAUTH2_INTROSPECTION_TOKEN_URL=http://my-oauth2-server/oauth2/token
			--------------------------------------------------------------

		- AUTHENTICATOR_OAUTH2_INTROSPECTION_SCOPE: If the OAuth 2.0 Token Introspection endpoint requires a certain OAuth 2.0 Scope
			in order to be accessed, you can set it using this environment variable. Use commas to define more than one OAuth 2.0 Scope.
			--------------------------------------------------------------
			Example: AUTHENTICATOR_OAUTH2_INTROSPECTION_SCOPE=scope-a,scope-b
			--------------------------------------------------------------


AUTHORIZERS
==============

- ORY Keto Warden Authorizer:
	- AUTHORIZER_KETO_URL: The URL of ORY Keto's URL. If the value is empty, then the ORY Keto Warden Authorizer
		will be disabled.
		--------------------------------------------------------------
		Example: AUTHORIZER_KETO_URL=http://keto-url/
		--------------------------------------------------------------


` + credentialsIssuer + `


OTHER CONTROLS
==============
- RULES_REFRESH_INTERVAL: ORY Oathkeeper stores rules in memory for faster access. This value sets the database refresh interval.
	Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	--------------------------------------------------------------
	Default: RULES_REFRESH_INTERVAL=5s
	--------------------------------------------------------------

- PROXY_SERVER_READ_TIMEOUT: The maximum duration for reading the entire request, including the body.
	Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	--------------------------------------------------------------
	Default: PROXY_SERVER_READ_TIMEOUT=5s
	--------------------------------------------------------------

- PROXY_SERVER_WRITE_TIMEOUT: The maximum duration before timing out writes of the response.
	Increase this parameter to prevent unexpected closing a client connection if an upstream request is executing more than 10 seconds. 
	Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	--------------------------------------------------------------
	Default: PROXY_SERVER_WRITE_TIMEOUT=10s
	--------------------------------------------------------------

- PROXY_SERVER_IDLE_TIMEOUT: The maximum amount of time to wait for any action of a request session, reading data or writing the response.
	Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	--------------------------------------------------------------
	Default: PROXY_SERVER_IDLE_TIMEOUT=120s
	--------------------------------------------------------------

` + corsMessage,
	Run: func(cmd *cobra.Command, args []string) {
		u, err := url.ParseRequestURI(viper.GetString("OATHKEEPER_API_URL"))
		if err != nil {
			logger.WithError(err).Fatalf(`Value from environment variable "OATHKEEPER_API_URL" is not a valid URL: %s`, err)
		}

		matcher := rule.NewHTTPMatcher(u)
		if err := matcher.Refresh(); err != nil {
			logger.WithError(err).Fatalln("Unable to refresh rules")
		}

		keyManager, err := keyManagerFactory(logger)
		if err != nil {
			logger.WithError(err).Fatalln("Unable to initialize the ID Token signing algorithm")
		}

		go refreshRules(matcher, 0)
		go refreshKeys(keyManager, 0)

		var authorizers = []authz.Authorizer{
			authz.NewAuthorizerAllow(),
			authz.NewAuthorizerDeny(),
		}

		if u := viper.GetString("AUTHORIZER_KETO_URL"); len(u) > 0 {
			authorizers = append(authorizers, authz.NewAuthorizerKetoEngineACPORY(urlx.ParseOrFatal(logger, u)))
		}

		authenticators, authorizers, credentialIssuers := handlerFactories(keyManager)
		eval := proxy.NewRequestHandler(logger, authenticators, authorizers, credentialIssuers)
		d := proxy.NewProxy(eval, logger, matcher)
		handler := &httputil.ReverseProxy{
			Director:  d.Director,
			Transport: d,
		}

		n := negroni.New()
		n.Use(negronilogrus.NewMiddlewareFromLogger(logger, "oathkeeper-proxy"))

		metrics := metricsx.New(cmd, logger,
			&metricsx.Options{
				Service:          "ory-oathkeeper",
				ClusterID:        metricsx.Hash(viper.GetString("DATABASE_URL")),
				IsDevelopment:    viper.GetString("DATABASE_URL") != "memory",
				WriteKey:         "MSx9A6YQ1qodnkzEFOv22cxOmOCJXMFa",
				WhitelistedPaths: []string{"/"},
				BuildVersion:     Version,
				BuildTime:        Commit,
				BuildHash:        Date,
			},
		)
		n.Use(metrics)

		n.UseHandler(handler)
		h := corsx.Initialize(n, logger, "")

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
			ReadTimeout:  viper.GetDuration("PROXY_SERVER_READ_TIMEOUT"),
			WriteTimeout: viper.GetDuration("PROXY_SERVER_WRITE_TIMEOUT"),
			IdleTimeout:  viper.GetDuration("PROXY_SERVER_IDLE_TIMEOUT"),
		})

		if err := graceful.Graceful(func() error {
			if cert != nil {
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
	},
}

func init() {
	serveCmd.AddCommand(proxyCmd)
}
