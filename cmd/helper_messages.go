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
	"fmt"
	"os"
)

var corsMessage = `CORS CONTROLS
==============
- CORS_ENABLED: Switch CORS support on (true) or off (false). Default is off (false).
	Example: CORS_ENABLED=true

- CORS_ALLOWED_ORIGINS: A list of origins (comma separated values) a cross-domain request can be executed from.
	If the special * value is present in the list, all origins will be allowed. An origin may contain a wildcard (*)
	to replace 0 or more characters (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penality.
	Only one wildcard can be used per origin. The default value is *.
	--------------------------------------------------------------
	Example: CORS_ALLOWED_ORIGINS=http://*.domain.com,http://*.domain2.com
	--------------------------------------------------------------

- CORS_ALLOWED_METHODS: A list of methods  (comma separated values) the client is allowed to use with cross-domain
	requests. Default value is simple methods (GET and POST).
	--------------------------------------------------------------
	Example: CORS_ALLOWED_METHODS=POST,GET,PUT
	--------------------------------------------------------------

- CORS_ALLOWED_CREDENTIALS: Indicates whether the request can include user credentials like cookies, HTTP authentication
	or client side SSL certificates.
	--------------------------------------------------------------
	Default: CORS_ALLOWED_CREDENTIALS=false
	Example: CORS_ALLOWED_CREDENTIALS=true
	--------------------------------------------------------------

- CORS_DEBUG: Debugging flag adds additional output to debug server side CORS issues.
	--------------------------------------------------------------
	Default: CORS_DEBUG=false
	Example: CORS_DEBUG=true
	--------------------------------------------------------------

- CORS_MAX_AGE: Indicates how long (in seconds) the results of a preflight request can be cached. The default is 0
	which stands for no max age.
	--------------------------------------------------------------
	Default: CORS_MAX_AGE=0
	Example: CORS_MAX_AGE=10
	--------------------------------------------------------------

- CORS_ALLOWED_HEADERS: A list of non simple headers (comma separated values) the client is allowed to use with
	cross-domain requests.

- CORS_EXPOSED_HEADERS: Indicates which headers (comma separated values) are safe to expose to the API of a
	CORS API specification.`

var databaseUrl = `- DATABASE_URL: A URL to a persistent backend. ORY Oathkeeper supports various backends:
  - Memory: If DATABASE_URL is "memory", data will be written to memory and is lost when you restart this instance.
	--------------------------------------------------------------
  	Example: DATABASE_URL=memory
	--------------------------------------------------------------

  - Postgres: If DATABASE_URL is a DSN starting with postgres:// PostgreSQL will be used as storage backend.
	--------------------------------------------------------------
	Example: DATABASE_URL=postgres://user:password@host:123/database
	--------------------------------------------------------------

	If PostgreSQL is not serving TLS, append ?sslmode=disable to the url:
	--------------------------------------------------------------
	DATABASE_URL=postgres://user:password@host:123/database?sslmode=disable
	--------------------------------------------------------------

  - MySQL: If DATABASE_URL is a DSN starting with mysql:// MySQL will be used as storage backend.
	--------------------------------------------------------------
	Example: DATABASE_URL=mysql://user:password@tcp(host:123)/database?parseTime=true
	--------------------------------------------------------------

	Be aware that the ?parseTime=true parameter is mandatory, or timestamps will not work.`

var credentialsIssuer = `CREDENTIALS ISSUERS
==============

- ID Token Credentials Issuer:

	- CREDENTIALS_ISSUER_ID_TOKEN_LIFESPAN: How long the ID token will be active. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
		--------------------------------------------------------------
		Default: CREDENTIALS_ISSUER_ID_TOKEN_LIFESPAN=10m
		--------------------------------------------------------------

	- CREDENTIALS_ISSUER_ID_TOKEN_ISSUER: Who issued the token - this will be the value of the "iss" claim in the ID Token.
		--------------------------------------------------------------
		Example: CREDENTIALS_ISSUER_ID_TOKEN_ISSUER=http://oathkeeper-url/
		--------------------------------------------------------------

	- CREDENTIALS_ISSUER_ID_TOKEN_JWK_REFRESH_INTERVAL: This value sets how often ORY Oathkeeper checks if a new
		key for signing is available. This is only required for strategies that fetch the key from a remote location.
		Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
		--------------------------------------------------------------
		Default: CREDENTIALS_ISSUER_ID_TOKEN_JWK_REFRESH_INTERVAL=5m
		--------------------------------------------------------------

	- CREDENTIALS_ISSUER_ID_TOKEN_ALGORITHM: The algorithm to be used for signing the ID Token. Supports HS256 (shared secret),
		"ORY-HYDRA" (uses ORY Hydra to create, store, and fetch RSA Keys for signing).
		--------------------------------------------------------------
		Default: CREDENTIALS_ISSUER_ID_TOKEN_ALGORITHM=HS256
		--------------------------------------------------------------
		Example: CREDENTIALS_ISSUER_ID_TOKEN_ALGORITHM=HS256
		Example: CREDENTIALS_ISSUER_ID_TOKEN_ALGORITHM=ORY-HYDRA
		--------------------------------------------------------------

		Depending on which algorithm you use, there are different configuration options necessary.

		HS256 ALGORITHM
		===============
			- CREDENTIALS_ISSUER_ID_TOKEN_HS256_SECRET: The shared secret to be used for the HS256 algorithm. The secret
				must be 32 characters long.
			--------------------------------------------------------------
			Example: CREDENTIALS_ISSUER_ID_TOKEN_HS256_SECRET=dYmTueb6zg8TphfZbOUpOewd0gt7u0SH
			--------------------------------------------------------------

		ORY-HYDRA ALGORITHM
		===============
			- CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_ADMIN_URL: The URL where ORY Hydra's Admin API is located.
				--------------------------------------------------------------
				Example: CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_ADMIN_URL=http://hydra-url/

			- CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_JWK_SET_ID: The JSON Web Key set identifier that will be used to create,
				store, and retrieve the JSON Web Key from ORY Hydra.
				--------------------------------------------------------------
				Default: CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_JWK_SET_ID=oathkeeper:id-token
				--------------------------------------------------------------

			If ORY Hydra's Admin API itself is protected with OAuth 2.0, you can provide the access credentials to perform
			an OAuth 2.0 Client Credentials flow before accessing ORY Hydra's APIs.

			These settings are usually not required and an optional! If you don't need this feature, leave them undefined.


				- CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_ID:  The ID of the OAuth 2.0 Client.
					--------------------------------------------------------------
					Example: CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_ID=my-client
					--------------------------------------------------------------

				- CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_SECRET: The secret of the OAuth 2.0 Client.
					--------------------------------------------------------------
					Example: CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_SECRET=my-secret
					--------------------------------------------------------------

				- CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_SCOPES: The OAuth 2.0 Scope the client should request.
					--------------------------------------------------------------
					Example: CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_CLIENT_SCOPES=foo,bar
					--------------------------------------------------------------

				- CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_PUBLIC_URL: The public URL where endpoint /oauth2/token is located.
					--------------------------------------------------------------
					Example: CREDENTIALS_ISSUER_ID_TOKEN_HYDRA_PUBLIC_URL=http://hydra-url/
					--------------------------------------------------------------
`

var tlsMessage = `
NOTE: configure TLS params consistently both as PATH or as string. If no TLS pair is set, HTTPS will be disabled and instead HTTP will be served.

- HTTPS_TLS_CERT_PATH: The path to the TLS certificate (pem encoded).
	Example: HTTPS_TLS_CERT_PATH=~/cert.pem

- HTTPS_TLS_KEY_PATH: The path to the TLS private key (pem encoded).
	Example: HTTPS_TLS_KEY_PATH=~/key.pem

- HTTPS_TLS_CERT: Base64 encoded (without padding) string of the TLS certificate (PEM encoded) to be used for HTTP over TLS (HTTPS).
	Example: HTTPS_TLS_CERT="-----BEGIN CERTIFICATE-----\nMIIDZTCCAk2gAwIBAgIEV5xOtDANBgkqhkiG9w0BAQ0FADA0MTIwMAYDVQQDDClP..."

- HTTPS_TLS_KEY: Base64 encoded (without padding) string of the private key (PEM encoded) to be used for HTTP over TLS (HTTPS).
	Example: HTTPS_TLS_KEY="-----BEGIN ENCRYPTED PRIVATE KEY-----\nMIIFDjBABgkqhkiG9w0BBQ0wMzAbBgkqhkiG9w0BBQwwDg..."
`

func fatalf(msg string, args ...interface{}) {
	fmt.Printf(msg+"\n", args...)
	os.Exit(1)
}
