---
id: configuration
title: Configuration
---

<!-- THIS FILE IS BEING AUTO-GENERATED. DO NOT MODIFY IT AS ALL CHANGES WILL BE OVERWRITTEN.
OPEN AN ISSUE IF YOU WOULD LIKE TO MAKE ADJUSTMENTS HERE AND MAINTAINERS WILL HELP YOU LOCATE THE RIGHT
FILE -->

If file `$HOME/.oathkeeper.yaml` exists, it will be used as a configuration file
which supports all configuration settings listed below.

You can load the config file from another source using the
`-c path/to/config.yaml` or `--config path/to/config.yaml` flag:
`oathkeeper --config path/to/config.yaml`.

Config files can be formatted as JSON, YAML and TOML. Some configuration values
support reloading without server restart. All configuration values can be set
using environment variables, as documented below.

To find out more about edge cases like setting string array values through
environmental variables head to the
[Configuring ORY services](https://www.ory.sh/docs/ecosystem/configuring)
section.

```yaml
## ORY Oathkeeper Configuration
#

## HTTP(s) ##
#
serve:
  ## HTTP REST API ##
  #
  api:
    ## Port ##
    #
    # The port to listen on.
    #
    # Default value: 4456
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_API_PORT=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_API_PORT=<value>
    #
    port: -8246268

    ## Host ##
    #
    # The network interface to listen on.
    #
    # Examples:
    # - localhost
    # - 127.0.0.1
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_API_HOST=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_API_HOST=<value>
    #
    host: localhost

    ## HTTP Timeouts ##
    #
    # Control the HTTP timeouts.
    #
    timeout:
      ## HTTP Read Timeout ##
      #
      # The maximum duration for reading the entire request, including the body.
      #
      # Default value: 5s
      #
      # Examples:
      # - 5s
      # - 5m
      # - 5h
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_API_TIMEOUT_READ=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_TIMEOUT_READ=<value>
      #
      read: 5h

      ## HTTP Write Timeout ##
      #
      # The maximum duration before timing out writes of the response. Increase this parameter to prevent unexpected closing a client connection if an upstream request is responding slowly.
      #
      # Default value: 120s
      #
      # Examples:
      # - 5s
      # - 5m
      # - 5h
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_API_TIMEOUT_WRITE=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_TIMEOUT_WRITE=<value>
      #
      write: 5s

      ## HTTP Idle Timeout ##
      #
      #  The maximum amount of time to wait for any action of a request session, reading data or writing the response.
      #
      # Default value: 120s
      #
      # Examples:
      # - 5s
      # - 5m
      # - 5h
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_API_TIMEOUT_IDLE=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_TIMEOUT_IDLE=<value>
      #
      idle: 5m

    ## Cross Origin Resource Sharing (CORS) ##
    #
    # Configure [Cross Origin Resource Sharing (CORS)](http://www.w3.org/TR/cors/) using the following options.
    #
    cors:
      ## Enable CORS ##
      #
      # If set to true, CORS will be enabled and preflight-requests (OPTION) will be answered.
      #
      # Default value: false
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_API_CORS_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_CORS_ENABLED=<value>
      #
      enabled: true

      ## Allowed Origins ##
      #
      # A list of origins a cross-domain request can be executed from. If the special * value is present in the list, all origins will be allowed. An origin may contain a wildcard (*) to replace 0 or more characters (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penality. Only one wildcard can be used per origin.
      #
      # Default value: *
      #
      # Examples:
      # - - https://example.com
      #   - https://*.example.com
      #   - https://*.foo.example.com
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_API_CORS_ALLOWED_ORIGINS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_CORS_ALLOWED_ORIGINS=<value>
      #
      allowed_origins:
        - '*'

      ## Allowed HTTP Methods ##
      #
      # A list of methods the client is allowed to use with cross-domain requests.
      #
      # Default value: GET,POST,PUT,PATCH,DELETE
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_API_CORS_ALLOWED_METHODS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_CORS_ALLOWED_METHODS=<value>
      #
      allowed_methods:
        - TRACE
        - GET

      ## Allowed Request HTTP Headers ##
      #
      # A list of non simple headers the client is allowed to use with cross-domain requests.
      #
      # Default value: Authorization,Content-Type
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_API_CORS_ALLOWED_HEADERS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_CORS_ALLOWED_HEADERS=<value>
      #
      allowed_headers:
        - consectetur elit ea quis
        - id Lorem

      ## Allowed Response HTTP Headers ##
      #
      # Indicates which headers are safe to expose to the API of a CORS API specification
      #
      # Default value: Content-Type
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_API_CORS_EXPOSED_HEADERS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_CORS_EXPOSED_HEADERS=<value>
      #
      exposed_headers:
        - sunt cillum
        - amet eiusmod

      ## Allow HTTP Credentials ##
      #
      # Indicates whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates.
      #
      # Default value: false
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_API_CORS_ALLOW_CREDENTIALS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_CORS_ALLOW_CREDENTIALS=<value>
      #
      allow_credentials: false

      ## Maximum Age ##
      #
      # Indicates how long (in seconds) the results of a preflight request can be cached. The default is 0 which stands for no max age.
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_API_CORS_MAX_AGE=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_CORS_MAX_AGE=<value>
      #
      max_age: -87069747

      ## Enable Debugging ##
      #
      # Set to true to debug server side CORS issues.
      #
      # Default value: false
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_API_CORS_DEBUG=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_CORS_DEBUG=<value>
      #
      debug: false

    ## HTTPS ##
    #
    # Configure HTTP over TLS (HTTPS). All options can also be set using environment variables by replacing dots (`.`) with underscores (`_`) and uppercasing the key. For example, `some.prefix.tls.key.path` becomes `export SOME_PREFIX_TLS_KEY_PATH`. If all keys are left undefined, TLS will be disabled.
    #
    tls:
      ## Private Key (PEM) ##
      #
      key:
        ## path ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SERVE_API_TLS_KEY_PATH=<value>
        # - Windows Command Line (CMD):
        #    > set SERVE_API_TLS_KEY_PATH=<value>
        #
        path: path/to/file.pem

        ## base64 ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SERVE_API_TLS_KEY_BASE64=<value>
        # - Windows Command Line (CMD):
        #    > set SERVE_API_TLS_KEY_BASE64=<value>
        #
        base64: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tXG5NSUlEWlRDQ0FrMmdBd0lCQWdJRVY1eE90REFOQmdr...

      ## TLS Certificate (PEM) ##
      #
      cert:
        ## path ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SERVE_API_TLS_CERT_PATH=<value>
        # - Windows Command Line (CMD):
        #    > set SERVE_API_TLS_CERT_PATH=<value>
        #
        path: path/to/file.pem

        ## base64 ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SERVE_API_TLS_CERT_BASE64=<value>
        # - Windows Command Line (CMD):
        #    > set SERVE_API_TLS_CERT_BASE64=<value>
        #
        base64: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tXG5NSUlEWlRDQ0FrMmdBd0lCQWdJRVY1eE90REFOQmdr...

  ## HTTP Reverse Proxy ##
  #
  proxy:
    ## Port ##
    #
    # The port to listen on.
    #
    # Default value: 4455
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_PROXY_PORT=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_PROXY_PORT=<value>
    #
    port: 41311124

    ## Host ##
    #
    # The network interface to listen on. Leave empty to listen on all interfaces.
    #
    # Examples:
    # - localhost
    # - 127.0.0.1
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_PROXY_HOST=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_PROXY_HOST=<value>
    #
    host: 127.0.0.1

    ## HTTP Timeouts ##
    #
    # Control the HTTP timeouts.
    #
    timeout:
      ## HTTP Read Timeout ##
      #
      # The maximum duration for reading the entire request, including the body.
      #
      # Default value: 5s
      #
      # Examples:
      # - 5s
      # - 5m
      # - 5h
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PROXY_TIMEOUT_READ=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PROXY_TIMEOUT_READ=<value>
      #
      read: 5s

      ## HTTP Write Timeout ##
      #
      # The maximum duration before timing out writes of the response. Increase this parameter to prevent unexpected closing a client connection if an upstream request is responding slowly.
      #
      # Default value: 120s
      #
      # Examples:
      # - 5s
      # - 5m
      # - 5h
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PROXY_TIMEOUT_WRITE=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PROXY_TIMEOUT_WRITE=<value>
      #
      write: 5m

      ## HTTP Idle Timeout ##
      #
      #  The maximum amount of time to wait for any action of a request session, reading data or writing the response.
      #
      # Default value: 120s
      #
      # Examples:
      # - 5s
      # - 5m
      # - 5h
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PROXY_TIMEOUT_IDLE=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PROXY_TIMEOUT_IDLE=<value>
      #
      idle: 120s

    ## Cross Origin Resource Sharing (CORS) ##
    #
    # Configure [Cross Origin Resource Sharing (CORS)](http://www.w3.org/TR/cors/) using the following options.
    #
    cors:
      ## Enable CORS ##
      #
      # If set to true, CORS will be enabled and preflight-requests (OPTION) will be answered.
      #
      # Default value: false
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PROXY_CORS_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PROXY_CORS_ENABLED=<value>
      #
      enabled: false

      ## Allowed Origins ##
      #
      # A list of origins a cross-domain request can be executed from. If the special * value is present in the list, all origins will be allowed. An origin may contain a wildcard (*) to replace 0 or more characters (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penality. Only one wildcard can be used per origin.
      #
      # Default value: *
      #
      # Examples:
      # - - https://example.com
      #   - https://*.example.com
      #   - https://*.foo.example.com
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PROXY_CORS_ALLOWED_ORIGINS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PROXY_CORS_ALLOWED_ORIGINS=<value>
      #
      allowed_origins:
        - '*'

      ## Allowed HTTP Methods ##
      #
      # A list of methods the client is allowed to use with cross-domain requests.
      #
      # Default value: GET,POST,PUT,PATCH,DELETE
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PROXY_CORS_ALLOWED_METHODS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PROXY_CORS_ALLOWED_METHODS=<value>
      #
      allowed_methods:
        - PUT
        - GET
        - PATCH

      ## Allowed Request HTTP Headers ##
      #
      # A list of non simple headers the client is allowed to use with cross-domain requests.
      #
      # Default value: Authorization,Content-Type
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PROXY_CORS_ALLOWED_HEADERS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PROXY_CORS_ALLOWED_HEADERS=<value>
      #
      allowed_headers:
        - labore ullamco

      ## Allowed Response HTTP Headers ##
      #
      # Indicates which headers are safe to expose to the API of a CORS API specification
      #
      # Default value: Content-Type
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PROXY_CORS_EXPOSED_HEADERS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PROXY_CORS_EXPOSED_HEADERS=<value>
      #
      exposed_headers:
        - nisi ea ut adipisicing eu

      ## Allow HTTP Credentials ##
      #
      # Indicates whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates.
      #
      # Default value: false
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PROXY_CORS_ALLOW_CREDENTIALS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PROXY_CORS_ALLOW_CREDENTIALS=<value>
      #
      allow_credentials: false

      ## Maximum Age ##
      #
      # Indicates how long (in seconds) the results of a preflight request can be cached. The default is 0 which stands for no max age.
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PROXY_CORS_MAX_AGE=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PROXY_CORS_MAX_AGE=<value>
      #
      max_age: -98873908

      ## Enable Debugging ##
      #
      # Set to true to debug server side CORS issues.
      #
      # Default value: false
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export SERVE_PROXY_CORS_DEBUG=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PROXY_CORS_DEBUG=<value>
      #
      debug: true

    ## HTTPS ##
    #
    # Configure HTTP over TLS (HTTPS). All options can also be set using environment variables by replacing dots (`.`) with underscores (`_`) and uppercasing the key. For example, `some.prefix.tls.key.path` becomes `export SOME_PREFIX_TLS_KEY_PATH`. If all keys are left undefined, TLS will be disabled.
    #
    tls:
      ## Private Key (PEM) ##
      #
      key:
        ## path ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SERVE_PROXY_TLS_KEY_PATH=<value>
        # - Windows Command Line (CMD):
        #    > set SERVE_PROXY_TLS_KEY_PATH=<value>
        #
        path: path/to/file.pem

        ## base64 ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SERVE_PROXY_TLS_KEY_BASE64=<value>
        # - Windows Command Line (CMD):
        #    > set SERVE_PROXY_TLS_KEY_BASE64=<value>
        #
        base64: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tXG5NSUlEWlRDQ0FrMmdBd0lCQWdJRVY1eE90REFOQmdr...

      ## TLS Certificate (PEM) ##
      #
      cert:
        ## path ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SERVE_PROXY_TLS_CERT_PATH=<value>
        # - Windows Command Line (CMD):
        #    > set SERVE_PROXY_TLS_CERT_PATH=<value>
        #
        path: path/to/file.pem

        ## base64 ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export SERVE_PROXY_TLS_CERT_BASE64=<value>
        # - Windows Command Line (CMD):
        #    > set SERVE_PROXY_TLS_CERT_BASE64=<value>
        #
        base64: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tXG5NSUlEWlRDQ0FrMmdBd0lCQWdJRVY1eE90REFOQmdr...

  ## Prometheus scraping endpoint ##
  #
  prometheus:
    ## Port ##
    #
    # The port to listen on.
    #
    # Default value: 9000
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_PROMETHEUS_PORT=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_PROMETHEUS_PORT=<value>
    #
    port: -93008656

    ## Host ##
    #
    # The network interface to listen on. Leave empty to listen on all interfaces.
    #
    # Examples:
    # - localhost
    # - 127.0.0.1
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_PROMETHEUS_HOST=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_PROMETHEUS_HOST=<value>
    #
    host: localhost

    ## Path ##
    #
    # The path to provide metrics on
    #
    # Default value: /metrics
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_PROMETHEUS_METRICS_PATH=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_PROMETHEUS_METRICS_PATH=<value>
    #
    metrics_path: ut id

    ## CollapsePaths ##
    #
    # When set to true the request label will include just the first segment of the request path
    #
    # Default value: true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export SERVE_PROMETHEUS_COLLAPSE_REQUEST_PATHS=<value>
    # - Windows Command Line (CMD):
    #    > set SERVE_PROMETHEUS_COLLAPSE_REQUEST_PATHS=<value>
    #
    collapse_request_paths: false

## Access Rules ##
#
# Configure access rules. All sub-keys support configuration reloading without restarting.
#
access_rules:
  ## Repositories ##
  #
  # Locations (list of URLs) where access rules should be fetched from on boot. It is expected that the documents at those locations return a JSON or YAML Array containing ORY Oathkeeper Access Rules:
  #
  # - If the URL Scheme is `file://`, the access rules (an array of access rules is expected) will be fetched from the local file system.
  # - If the URL Scheme is `inline://`, the access rules (an array of access rules is expected) are expected to be a base64 encoded (with padding!) JSON/YAML string (base64_encode(`[{"id":"foo-rule","authenticators":[....]}]`)).
  # - If the URL Scheme is `http://` or `https://`, the access rules (an array of access rules is expected) will be fetched from the provided HTTP(s) location.
  #
  # Examples:
  # - - file://path/to/rules.json
  #   - inline://W3siaWQiOiJmb28tcnVsZSIsImF1dGhlbnRpY2F0b3JzIjpbXX1d
  #   - https://path-to-my-rules/rules.json
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export ACCESS_RULES_REPOSITORIES=<value>
  # - Windows Command Line (CMD):
  #    > set ACCESS_RULES_REPOSITORIES=<value>
  #
  repositories:
    - file://path/to/rules.json
    - inline://W3siaWQiOiJmb28tcnVsZSIsImF1dGhlbnRpY2F0b3JzIjpbXX1d
    - https://path-to-my-rules/rules.json

  ## Matching strategy ##
  #
  # This an optional field describing matching strategy. Currently supported values are 'glob' and 'regexp'.
  #
  # Default value: regexp
  #
  # Examples:
  # - glob
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export ACCESS_RULES_MATCHING_STRATEGY=<value>
  # - Windows Command Line (CMD):
  #    > set ACCESS_RULES_MATCHING_STRATEGY=<value>
  #
  matching_strategy: glob

## Authenticators ##
#
# For more information on authenticators head over to: https://www.ory.sh/oathkeeper/docs/pipeline/authn
#
authenticators:
  ## Anonymous ##
  #
  # The [`anonymous` authenticator](https://www.ory.sh/oathkeeper/docs/pipeline/authn#anonymous).
  #
  anonymous:
    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export AUTHENTICATORS_ANONYMOUS_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set AUTHENTICATORS_ANONYMOUS_ENABLED=<value>
    #
    enabled: true

    ## Anonymous Authenticator Configuration ##
    #
    # This section is optional when the authenticator is disabled.
    #
    config:
      ## Anonymous Subject ##
      #
      # Sets the anonymous username.
      #
      # Default value: anonymous
      #
      # Examples:
      # - guest
      # - anon
      # - unknown
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHENTICATORS_ANONYMOUS_CONFIG_SUBJECT=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHENTICATORS_ANONYMOUS_CONFIG_SUBJECT=<value>
      #
      subject: anonymous

  ## No Operation (noop) ##
  #
  # The [`noop` authenticator](https://www.ory.sh/oathkeeper/docs/pipeline/authn#noop).
  #
  noop:
    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export AUTHENTICATORS_NOOP_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set AUTHENTICATORS_NOOP_ENABLED=<value>
    #
    enabled: true

  ## Unauthorized ##
  #
  # The [`unauthorized` authenticator](https://www.ory.sh/oathkeeper/docs/pipeline/authn#unauthorized).
  #
  unauthorized:
    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export AUTHENTICATORS_UNAUTHORIZED_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set AUTHENTICATORS_UNAUTHORIZED_ENABLED=<value>
    #
    enabled: false

  ## Cookie Session ##
  #
  # The [`cookie_session` authenticator](https://www.ory.sh/oathkeeper/docs/pipeline/authn#cookie_session).
  #
  cookie_session:
    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export AUTHENTICATORS_COOKIE_SESSION_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set AUTHENTICATORS_COOKIE_SESSION_ENABLED=<value>
    #
    enabled: true

  ## JSON Web Token (jwt) ##
  #
  # The [`jwt` authenticator](https://www.ory.sh/oathkeeper/docs/pipeline/authn#jwt).
  #
  jwt:
    ## config ##
    #
    config:
      ## jwks_urls ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHENTICATORS_JWT_CONFIG_JWKS_URLS=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHENTICATORS_JWT_CONFIG_JWKS_URLS=<value>
      #
      jwks_urls:
        - https://my-website.com/.well-known/jwks.json
        - https://my-other-website.com/.well-known/jwks.json
        - file://path/to/local/jwks.json

      ## required_scope ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHENTICATORS_JWT_CONFIG_REQUIRED_SCOPE=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHENTICATORS_JWT_CONFIG_REQUIRED_SCOPE=<value>
      #
      required_scope:
        - esse cupidatat
        - sint nulla sunt
        - quis qui aliquip Excepteur adipisicing
        - quis eiusmod anim
        - aliqua laboris id sint

      ## target_audience ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHENTICATORS_JWT_CONFIG_TARGET_AUDIENCE=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHENTICATORS_JWT_CONFIG_TARGET_AUDIENCE=<value>
      #
      target_audience:
        - minim adipisicing fugiat aliquip

      ## trusted_issuers ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHENTICATORS_JWT_CONFIG_TRUSTED_ISSUERS=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHENTICATORS_JWT_CONFIG_TRUSTED_ISSUERS=<value>
      #
      trusted_issuers:
        - commodo eu pariatur
        - labore est dolor sint

      ## allowed_algorithms ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHENTICATORS_JWT_CONFIG_ALLOWED_ALGORITHMS=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHENTICATORS_JWT_CONFIG_ALLOWED_ALGORITHMS=<value>
      #
      allowed_algorithms:
        - cillum quis
        - sit deserunt ut in
        - dolore ad
        - enim Lorem labore Duis laborum

      ## scope_strategy ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHENTICATORS_JWT_CONFIG_SCOPE_STRATEGY=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHENTICATORS_JWT_CONFIG_SCOPE_STRATEGY=<value>
      #
      scope_strategy: exact

      ## token_from ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHENTICATORS_JWT_CONFIG_TOKEN_FROM=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHENTICATORS_JWT_CONFIG_TOKEN_FROM=<value>
      #
      token_from: null

    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export AUTHENTICATORS_JWT_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set AUTHENTICATORS_JWT_ENABLED=<value>
    #
    enabled: false

  ## OAuth 2.0 Client Credentials ##
  #
  # The [`oauth2_client_credentials` authenticator](https://www.ory.sh/oathkeeper/docs/pipeline/authn#oauth2_client_credentials).
  #
  oauth2_client_credentials:
    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export AUTHENTICATORS_OAUTH2_CLIENT_CREDENTIALS_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set AUTHENTICATORS_OAUTH2_CLIENT_CREDENTIALS_ENABLED=<value>
    #
    enabled: false

  ## OAuth 2.0 Token Introspection ##
  #
  # The [`oauth2_introspection` authenticator](https://www.ory.sh/oathkeeper/docs/pipeline/authn#oauth2_introspection).
  #
  oauth2_introspection:
    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_ENABLED=<value>
    #
    enabled: false

## Error Handling ##
#
errors:
  ## Error Handling Fallback ##
  #
  # This array defines how to handle errors when no "when" clause matches. If you have, for example, enabled redirect and json in your access rule, you could tell ORY Oathkeeper to try sending JSON if the request does not match the access rule definition
  #
  # Default value: json
  #
  # Examples:
  # - - redirect
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export ERRORS_FALLBACK=<value>
  # - Windows Command Line (CMD):
  #    > set ERRORS_FALLBACK=<value>
  #
  fallback:
    - redirect

  ## Individual Error Handler Configuration ##
  #
  handlers:
    ## HTTP WWW-Authenticate Handler ##
    #
    # Responds with the WWW-Authenticate HTTP Response
    #
    www_authenticate:
      ## config ##
      #
      config:
        ## realm ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export ERRORS_HANDLERS_WWW_AUTHENTICATE_CONFIG_REALM=<value>
        # - Windows Command Line (CMD):
        #    > set ERRORS_HANDLERS_WWW_AUTHENTICATE_CONFIG_REALM=<value>
        #
        realm: dolor sint

        ## when ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export ERRORS_HANDLERS_WWW_AUTHENTICATE_CONFIG_WHEN=<value>
        # - Windows Command Line (CMD):
        #    > set ERRORS_HANDLERS_WWW_AUTHENTICATE_CONFIG_WHEN=<value>
        #
        when:
          - error:
              - unauthorized
              - forbidden
              - unauthorized
              - unauthorized
            request:
              cidr:
                - mollit
                - sint
              header:
                content_type: []
                accept: []
          - error:
              - internal_server_error
              - forbidden
            request:
              cidr:
                - Ut ad
              header:
                content_type: []
                accept: []
          - error:
              - internal_server_error
              - not_found
            request:
              cidr:
                - nulla voluptate sint
                - laboris culpa elit sit et
                - do sed veniam culpa adipisicing
                - sunt irure
                - enim Lorem do
              header:
                content_type: []
                accept: []

      ## Enabled ##
      #
      # En-/disables this component.
      #
      # Default value: false
      #
      # Examples:
      # - true
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export ERRORS_HANDLERS_WWW_AUTHENTICATE_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set ERRORS_HANDLERS_WWW_AUTHENTICATE_ENABLED=<value>
      #
      enabled: false

    ## HTTP Redirect Error Handler ##
    #
    # Responds with a 301/302 HTTP redirect.
    #
    redirect:
      ## Enabled ##
      #
      # En-/disables this component.
      #
      # Default value: false
      #
      # Examples:
      # - true
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export ERRORS_HANDLERS_REDIRECT_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set ERRORS_HANDLERS_REDIRECT_ENABLED=<value>
      #
      enabled: false

    ## JSON Error Handler ##
    #
    # Responds with a JSON error response
    #
    # Default value: [object Object]
    #
    json:
      ## Enabled ##
      #
      # En-/disables this component.
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export ERRORS_HANDLERS_JSON_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set ERRORS_HANDLERS_JSON_ENABLED=<value>
      #
      enabled: false

## Authorizers ##
#
# For more information on authorizers head over to: https://www.ory.sh/oathkeeper/docs/pipeline/authz
#
authorizers:
  ## Allow ##
  #
  # The [`allow` authorizer](https://www.ory.sh/oathkeeper/docs/pipeline/authz#allow).
  #
  allow:
    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export AUTHORIZERS_ALLOW_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set AUTHORIZERS_ALLOW_ENABLED=<value>
    #
    enabled: false

  ## Deny ##
  #
  # The [`deny` authorizer](https://www.ory.sh/oathkeeper/docs/pipeline/authz#allow).
  #
  deny:
    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export AUTHORIZERS_DENY_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set AUTHORIZERS_DENY_ENABLED=<value>
    #
    enabled: true

  ## ORY Keto Access Control Policies Engine ##
  #
  # The [`keto_engine_acp_ory` authorizer](https://www.ory.sh/oathkeeper/docs/pipeline/authz#keto_engine_acp_ory).
  #
  keto_engine_acp_ory:
    ## config ##
    #
    config:
      ## base_url ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_BASE_URL=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_BASE_URL=<value>
      #
      base_url: http://my-keto/

      ## required_action ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_REQUIRED_ACTION=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_REQUIRED_ACTION=<value>
      #
      required_action: elit occaecat ea veniam voluptate

      ## required_resource ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_REQUIRED_RESOURCE=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_REQUIRED_RESOURCE=<value>
      #
      required_resource: voluptate magna in consectetur cillum

      ## subject ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_SUBJECT=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_SUBJECT=<value>
      #
      subject: labore in

      ## flavor ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_FLAVOR=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_FLAVOR=<value>
      #
      flavor: qui

    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export AUTHORIZERS_KETO_ENGINE_ACP_ORY_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set AUTHORIZERS_KETO_ENGINE_ACP_ORY_ENABLED=<value>
    #
    enabled: false

  ## Remote ##
  #
  # The [`remote` authorizer](https://www.ory.sh/oathkeeper/docs/pipeline/authz#remote).
  #
  remote:
    ## config ##
    #
    config:
      ## remote ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHORIZERS_REMOTE_CONFIG_REMOTE=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHORIZERS_REMOTE_CONFIG_REMOTE=<value>
      #
      remote: https://host/path

    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export AUTHORIZERS_REMOTE_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set AUTHORIZERS_REMOTE_ENABLED=<value>
    #
    enabled: false

  ## Remote JSON ##
  #
  # The [`remote_json` authorizer](https://www.ory.sh/oathkeeper/docs/pipeline/authz#remote_json).
  #
  remote_json:
    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export AUTHORIZERS_REMOTE_JSON_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set AUTHORIZERS_REMOTE_JSON_ENABLED=<value>
    #
    enabled: false

## Mutators ##
#
# For more information on mutators head over to: https://www.ory.sh/oathkeeper/docs/pipeline/mutator
#
mutators:
  ## No Operation (noop) ##
  #
  # The [`noop` mutator](https://www.ory.sh/oathkeeper/docs/pipeline/mutator#noop).
  #
  noop:
    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export MUTATORS_NOOP_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set MUTATORS_NOOP_ENABLED=<value>
    #
    enabled: false

  ## HTTP Cookie ##
  #
  # The [`cookie` mutator](https://www.ory.sh/oathkeeper/docs/pipeline/mutator#cookie).
  #
  cookie:
    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export MUTATORS_COOKIE_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set MUTATORS_COOKIE_ENABLED=<value>
    #
    enabled: true

  ## HTTP Header ##
  #
  # The [`header` mutator](https://www.ory.sh/oathkeeper/docs/pipeline/mutator#header).
  #
  header:
    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export MUTATORS_HEADER_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set MUTATORS_HEADER_ENABLED=<value>
    #
    enabled: false

  ## Hydrator ##
  #
  # The [`hydrator` mutator](https://www.ory.sh/oathkeeper/docs/pipeline/mutator#hydrator).
  #
  hydrator:
    ## config ##
    #
    config:
      ## api ##
      #
      api:
        ## url ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export MUTATORS_HYDRATOR_CONFIG_API_URL=<value>
        # - Windows Command Line (CMD):
        #    > set MUTATORS_HYDRATOR_CONFIG_API_URL=<value>
        #
        url: http://bMfa.vqqsbZC6M+pl8c9PoFx4h6amQIlaRZmGcTjX0wD-FTWHHy.oO9XRIxRJDb2MjVqtFLhQd8APcQ

        ## auth ##
        #
        auth:
          ## basic ##
          #
          basic:
            ## username ##
            #
            # Set this value using environment variables on
            # - Linux/macOS:
            #    $ export MUTATORS_HYDRATOR_CONFIG_API_AUTH_BASIC_USERNAME=<value>
            # - Windows Command Line (CMD):
            #    > set MUTATORS_HYDRATOR_CONFIG_API_AUTH_BASIC_USERNAME=<value>
            #
            username: nostrud et

            ## password ##
            #
            # Set this value using environment variables on
            # - Linux/macOS:
            #    $ export MUTATORS_HYDRATOR_CONFIG_API_AUTH_BASIC_PASSWORD=<value>
            # - Windows Command Line (CMD):
            #    > set MUTATORS_HYDRATOR_CONFIG_API_AUTH_BASIC_PASSWORD=<value>
            #
            password: veniam anim reprehenderit in et

        ## retry ##
        #
        retry:
          ## give_up_after ##
          #
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export MUTATORS_HYDRATOR_CONFIG_API_RETRY_GIVE_UP_AFTER=<value>
          # - Windows Command Line (CMD):
          #    > set MUTATORS_HYDRATOR_CONFIG_API_RETRY_GIVE_UP_AFTER=<value>
          #
          give_up_after: 5606450m

          ## max_delay ##
          #
          # Set this value using environment variables on
          # - Linux/macOS:
          #    $ export MUTATORS_HYDRATOR_CONFIG_API_RETRY_MAX_DELAY=<value>
          # - Windows Command Line (CMD):
          #    > set MUTATORS_HYDRATOR_CONFIG_API_RETRY_MAX_DELAY=<value>
          #
          max_delay: 852m

      ## cache ##
      #
      cache:
        ## enabled ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export MUTATORS_HYDRATOR_CONFIG_CACHE_ENABLED=<value>
        # - Windows Command Line (CMD):
        #    > set MUTATORS_HYDRATOR_CONFIG_CACHE_ENABLED=<value>
        #
        enabled: false

        ## ttl ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export MUTATORS_HYDRATOR_CONFIG_CACHE_TTL=<value>
        # - Windows Command Line (CMD):
        #    > set MUTATORS_HYDRATOR_CONFIG_CACHE_TTL=<value>
        #
        ttl: 8070709h

    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export MUTATORS_HYDRATOR_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set MUTATORS_HYDRATOR_ENABLED=<value>
    #
    enabled: true

  ## ID Token (JSON Web Token) ##
  #
  # The [`id_token` mutator](https://www.ory.sh/oathkeeper/docs/pipeline/mutator#id_token).
  #
  id_token:
    ## Enabled ##
    #
    # En-/disables this component.
    #
    # Default value: false
    #
    # Examples:
    # - true
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export MUTATORS_ID_TOKEN_ENABLED=<value>
    # - Windows Command Line (CMD):
    #    > set MUTATORS_ID_TOKEN_ENABLED=<value>
    #
    enabled: false

## Log ##
#
# Configure logging using the following options. Logging will always be sent to stdout and stderr.
#
log:
  ## Level ##
  #
  # Debug enables stack traces on errors. Can also be set using environment variable LOG_LEVEL.
  #
  # Default value: info
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export LOG_LEVEL=<value>
  # - Windows Command Line (CMD):
  #    > set LOG_LEVEL=<value>
  #
  level: debug

  ## Format ##
  #
  # The log format can either be text or JSON.
  #
  # Default value: text
  #
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export LOG_FORMAT=<value>
  # - Windows Command Line (CMD):
  #    > set LOG_FORMAT=<value>
  #
  format: text

## Profiling ##
#
# Enables CPU or memory profiling if set. For more details on profiling Go programs read [Profiling Go Programs](https://blog.golang.org/profiling-go-programs).
#
# Set this value using environment variables on
# - Linux/macOS:
#    $ export PROFILING=<value>
# - Windows Command Line (CMD):
#    > set PROFILING=<value>
#
profiling: cpu
```
