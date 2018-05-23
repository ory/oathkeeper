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

	"strings"

	"github.com/meatballhat/negroni-logrus"
	"github.com/ory/fosite"
	"github.com/ory/graceful"
	"github.com/ory/keto/sdk/go/keto"
	"github.com/ory/metrics-middleware"
	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rsakey"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/sdk/go/oathkeeper"
	"github.com/rs/cors"
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

- OATHKEEPER_API_URL: The URL of the Oathkeeper REST API
	Example: OATHKEEPER_API_URL=https://api.oathkeeper.mydomain.com/


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


AUTHENTICATORS
==============

- OAuth 2.0 Client Credentials Authenticator:
	- AUTHENTICATOR_OAUTH2_CLIENT_CREDENTIALS_TOKEN_URL: Sets the OAuth 2.0 Token URL that should be used to check if the provided credentials are valid or not.
		Example: AUTHENTICATOR_OAUTH2_CLIENT_CREDENTIALS_TOKEN_URL=http://my-oauth2-server/oauth2/token

- OAuth 2.0 Token Introspection Authenticator:
	- AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_ID: The OAuth 2.0 Client ID the client that performs the OAuth 2.0 Token Introspection. The OAuth 2.0 Token Introspection
    	endpoint is typically protected and requires a valid OAuth 2.0 Client in order to check if a token is valid or not.
		Example: AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_ID=my-client-id

	- AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_SECRET:T he OAuth 2.0 Client Secret of the client that performs the OAuth 2.0 Token Introspection.
		Example: AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_ID=my-client-secret

	- AUTHENTICATOR_OAUTH2_INTROSPECTION_TOKEN_URL: The OAuth 2.0 Token URL.
		Example: AUTHENTICATOR_OAUTH2_INTROSPECTION_TOKEN_URL=http://my-oauth2-server/oauth2/token

	- AUTHENTICATOR_OAUTH2_INTROSPECTION_INTROSPECT_URL: The OAuth 2.0 Token Introspection URL.
		Example: AUTHENTICATOR_OAUTH2_INTROSPECTION_INTROSPECT_URL=http://my-oauth2-server/oauth2/introspect

	- AUTHENTICATOR_OAUTH2_INTROSPECTION_SCOPE: If the OAuth 2.0 Token Introspection endpoint requires a certain OAuth 2.0 Scope
    	in order to be accessed, you can set it using this environment variable. Use commas to define more than one OAuth 2.0 Scope.
		Example: AUTHENTICATOR_OAUTH2_INTROSPECTION_SCOPE=scope-a,scope-b


AUTHORIZERS
==============

- ORY Keto Warden Authorizer:
	- AUTHORIZER_KETO_WARDEN_KETO_URL: The URL of ORY Keto's URL.
		Example: AUTHORIZER_KETO_WARDEN_KETO_URL=http://keto-url/


CREDENTIALS ISSUERS
==============

- ID Token Credentials Issuer:
	- CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_URL: The URL where ORY Hydra is located.
		Example: CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_URL=http://hydra-url/

	- CREDENTIALS_ISSUER_ID_TOKEN_JWK_REFRESH_INTERVAL: ORY Oathkeeper stores JSON Web Keys for ID Token signing in memory. This value sets the refresh interval.
		Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
		Default: CREDENTIALS_ISSUER_ID_TOKEN_JWK_REFRESH_INTERVAL=5m

	- CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_JWK_SET_ID: The JSON Web Key set identifier that will be used to create, store, and retrieve the JSON Web Key from ORY Hydra.
		Default: CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_JWK_SET_ID=oathkeeper:id-token

	- CREDENTIALS_ISSUER_ID_TOKEN_LIFESPAN: How long the ID token will be active. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
		Default: CREDENTIALS_ISSUER_ID_TOKEN_LIFESPAN=10m

	- CREDENTIALS_ISSUER_ID_TOKEN_ISSUER: Who issued the token - this will be the value of the "iss" claim in the ID Token.
		Example: CREDENTIALS_ISSUER_ID_TOKEN_ISSUER=http://oathkeeper-url/


OTHER CONTROLS
==============
- RULES_REFRESH_INTERVAL: ORY Oathkeeper stores rules in memory for faster access. This value sets the database refresh interval.
	Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	Default: RULES_REFRESH_INTERVAL=5s
` + corsMessage,
	Run: func(cmd *cobra.Command, args []string) {
		oathkeeperSdk := oathkeeper.NewSDK(viper.GetString("OATHKEEPER_API_URL"))
		hydraSdk := getHydraSDK()
		ketoSdk, err := keto.NewCodeGenSDK(&keto.Configuration{
			EndpointURL: viper.GetString("AUTHORIZER_KETO_WARDEN_KETO_URL"),
		})
		if err != nil {
			logger.WithError(err).Fatal("Unable to initialize the ORY Keto SDK")
		}

		matcher := rule.NewHTTPMatcher(oathkeeperSdk)
		if err := matcher.Refresh(); err != nil {
			logger.WithError(err).Fatalln("Unable to refresh rules")
		}

		keyManager := &rsakey.HydraManager{
			SDK: hydraSdk,
			Set: viper.GetString("CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_JWK_SET_ID"),
		}

		go refreshRules(matcher, 0)
		go refreshKeys(keyManager, 0)

		eval := proxy.NewRequestHandler(
			logger,
			[]proxy.Authenticator{
				proxy.NewAuthenticatorNoOp(),
				proxy.NewAuthenticatorAnonymous(viper.GetString("AUTHENTICATOR_ANONYMOUS_USERNAME")),
				proxy.NewAuthenticatorOAuth2Introspection(
					viper.GetString("AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_ID"),
					viper.GetString("AUTHENTICATOR_OAUTH2_INTROSPECTION_CLIENT_SECRET"),
					viper.GetString("AUTHENTICATOR_OAUTH2_INTROSPECTION_TOKEN_URL"),
					viper.GetString("AUTHENTICATOR_OAUTH2_INTROSPECTION_INTROSPECT_URL"),
					strings.Split(viper.GetString("AUTHENTICATOR_OAUTH2_INTROSPECTION_SCOPE"), ","),
					fosite.WildcardScopeStrategy,
				),
				proxy.NewAuthenticatorOAuth2ClientCredentials(
					viper.GetString("AUTHENTICATOR_OAUTH2_CLIENT_CREDENTIALS_TOKEN_URL"),
				),
			},
			[]proxy.Authorizer{
				proxy.NewAuthorizerAllow(),
				proxy.NewAuthorizerDeny(),
				proxy.NewAuthorizerKetoWarden(ketoSdk),
			},
			[]proxy.CredentialsIssuer{
				proxy.NewCredentialsIssuerNoOp(),
				proxy.NewCredentialsIssuerIDToken(
					keyManager,
					logger,
					viper.GetDuration("CREDENTIALS_ISSUER_ID_TOKEN_LIFESPAN"),
					viper.GetString("CREDENTIALS_ISSUER_ID_TOKEN_ISSUER"),
				),
			},
		)
		d := proxy.NewProxy(eval, logger, matcher)
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
