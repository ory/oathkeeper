---
id: configuration
title: Configuration
---

<!-- THIS FILE IS BEING AUTO-GENERATED. DO NOT MODIFY IT AS ALL CHANGES WILL BE OVERWRITTEN.
OPEN AN ISSUE IF YOU WOULD LIKE TO MAKE ADJUSTMENTS HERE AND MAINTAINERS WILL HELP YOU LOCATE THE RIGHT
FILE -->

If file `$HOME/.oathkeeper.yaml` exists, it will be used as a configuration file which supports all
configuration settings listed below.

You can load the config file from another source using the `-c path/to/config.yaml` or `--config path/to/config.yaml`
flag: `oathkeeper --config path/to/config.yaml`.

Config files can be formatted as JSON, YAML and TOML. Some configuration values support reloading without server restart.
All configuration values can be set using environment variables, as documented below.

To find out more about edge cases like setting string array values through environmental variables head to the
[Configuring ORY services](https://www.ory.sh/docs/ecosystem/configuring) section.

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
    port: 84865020

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
      #    $ export SERVE_API_TIMEOUT_READ=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_TIMEOUT_READ=<value>
      #
      read: 5m

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
      write: 5h

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
      #    $ export SERVE_API_CORS_ENABLED=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_CORS_ENABLED=<value>
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
      #    $ export SERVE_API_CORS_ALLOWED_ORIGINS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_API_CORS_ALLOWED_ORIGINS=<value>
      #
      allowed_origins:
        - "*"

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
        - PATCH
        - DELETE

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
        - sed nisi ex proident non
        - veniam velit ut proident consequat

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
        - adipisicing laborum Lorem sed aute
        - voluptate Duis reprehenderit
        - aliquip Ut dolore pariatur cillum
        - anim in nisi ipsum velit

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
      allow_credentials: true

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
      max_age: -14039419

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
    port: 3425646

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
      write: 5h

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
      idle: 5s

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
      #    $ export SERVE_PROXY_CORS_ALLOWED_ORIGINS=<value>
      # - Windows Command Line (CMD):
      #    > set SERVE_PROXY_CORS_ALLOWED_ORIGINS=<value>
      #
      allowed_origins:
        - "*"

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
        - GET
        - CONNECT
        - PATCH
        - PUT

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
        - incididunt cillum ipsum
        - sit culpa cillum
        - ex commodo proident in

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
        - fugiat eu
        - deserunt consequat sed

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
      max_age: 26387686

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
    port: -83436648

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
    host: ""

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
    metrics_path: ut culpa nulla Ut

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
    collapse_request_paths: true

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
  # One of:
  # - glob
  # - regexp
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
  matching_strategy: regexp

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
      subject: guest

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
    enabled: false

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
    enabled: true

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
    enabled: true

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
    enabled: true

  ## OAuth 2.0 Token Introspection ##
  #
  # The [`oauth2_introspection` authenticator](https://www.ory.sh/oathkeeper/docs/pipeline/authn#oauth2_introspection).
  #
  oauth2_introspection:
    
    ## config ##
    #
    config:
      
      ## introspection_url ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_INTROSPECTION_URL=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_INTROSPECTION_URL=<value>
      #
      introspection_url: https://my-website.com/oauth2/introspection

      ## scope_strategy ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_SCOPE_STRATEGY=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_SCOPE_STRATEGY=<value>
      #
      scope_strategy: exact

      ## pre_authorization ##
      #
      pre_authorization:
        
        ## client_id ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_PRE_AUTHORIZATION_CLIENT_ID=<value>
        # - Windows Command Line (CMD):
        #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_PRE_AUTHORIZATION_CLIENT_ID=<value>
        #
        client_id: ea et laboris cillum

        ## client_secret ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_PRE_AUTHORIZATION_CLIENT_SECRET=<value>
        # - Windows Command Line (CMD):
        #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_PRE_AUTHORIZATION_CLIENT_SECRET=<value>
        #
        client_secret: aliquip non

        ## token_url ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_PRE_AUTHORIZATION_TOKEN_URL=<value>
        # - Windows Command Line (CMD):
        #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_PRE_AUTHORIZATION_TOKEN_URL=<value>
        #
        token_url: http://ZyJWHYLMFGmekjRKr.tssB+,JCGIbCNGJ6riBrnHbk+XukF-VlXwkiqevQzx9Q03

        ## enabled ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_PRE_AUTHORIZATION_ENABLED=<value>
        # - Windows Command Line (CMD):
        #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_PRE_AUTHORIZATION_ENABLED=<value>
        #
        enabled: true

        ## scope ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_PRE_AUTHORIZATION_SCOPE=<value>
        # - Windows Command Line (CMD):
        #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_PRE_AUTHORIZATION_SCOPE=<value>
        #
        scope:
          - foo
          - bar

      ## required_scope ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_REQUIRED_SCOPE=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_REQUIRED_SCOPE=<value>
      #
      required_scope:
        - aliqua enim magna pariatur Ut
        - culpa adipisicing sunt

      ## target_audience ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_TARGET_AUDIENCE=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_TARGET_AUDIENCE=<value>
      #
      target_audience:
        - aute

      ## trusted_issuers ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_TRUSTED_ISSUERS=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_TRUSTED_ISSUERS=<value>
      #
      trusted_issuers:
        - eiusmod
        - aute dolore elit anim Excepteur
        - Ut incididunt in laborum

      ## token_from ##
      #
      token_from:
        
        ## query_parameter ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_TOKEN_FROM_QUERY_PARAMETER=<value>
        # - Windows Command Line (CMD):
        #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_TOKEN_FROM_QUERY_PARAMETER=<value>
        #
        query_parameter: eiusmod adipisicing veniam ipsum exercitation

      ## retry ##
      #
      retry:
        
        ## give_up_after ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_RETRY_GIVE_UP_AFTER=<value>
        # - Windows Command Line (CMD):
        #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_RETRY_GIVE_UP_AFTER=<value>
        #
        give_up_after: 458230s

        ## max_delay ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_RETRY_MAX_DELAY=<value>
        # - Windows Command Line (CMD):
        #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_RETRY_MAX_DELAY=<value>
        #
        max_delay: 8h

      ## cache ##
      #
      cache:
        
        ## enabled ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_CACHE_ENABLED=<value>
        # - Windows Command Line (CMD):
        #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_CACHE_ENABLED=<value>
        #
        enabled: true

        ## ttl ##
        #
        # Set this value using environment variables on
        # - Linux/macOS:
        #    $ export AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_CACHE_TTL=<value>
        # - Windows Command Line (CMD):
        #    > set AUTHENTICATORS_OAUTH2_INTROSPECTION_CONFIG_CACHE_TTL=<value>
        #
        ttl: 5s

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
    enabled: true

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
      enabled: true

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
      required_action: esse aliqua

      ## required_resource ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_REQUIRED_RESOURCE=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_REQUIRED_RESOURCE=<value>
      #
      required_resource: dolor

      ## subject ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_SUBJECT=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_SUBJECT=<value>
      #
      subject: aliquip dolor enim ut do

      ## flavor ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_FLAVOR=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHORIZERS_KETO_ENGINE_ACP_ORY_CONFIG_FLAVOR=<value>
      #
      flavor: anim voluptate velit in

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
    
    ## config ##
    #
    config:
      
      ## remote ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHORIZERS_REMOTE_JSON_CONFIG_REMOTE=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHORIZERS_REMOTE_JSON_CONFIG_REMOTE=<value>
      #
      remote: https://host/path

      ## payload ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export AUTHORIZERS_REMOTE_JSON_CONFIG_PAYLOAD=<value>
      # - Windows Command Line (CMD):
      #    > set AUTHORIZERS_REMOTE_JSON_CONFIG_PAYLOAD=<value>
      #
      payload: '{"subject":"{{ .Subject }}"}'

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
    
    ## config ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export MUTATORS_COOKIE_CONFIG=<value>
    # - Windows Command Line (CMD):
    #    > set MUTATORS_COOKIE_CONFIG=<value>
    #
    config: {}

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
    
    ## config ##
    #
    # Set this value using environment variables on
    # - Linux/macOS:
    #    $ export MUTATORS_HEADER_CONFIG=<value>
    # - Windows Command Line (CMD):
    #    > set MUTATORS_HEADER_CONFIG=<value>
    #
    config: {}

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
    enabled: false

  ## ID Token (JSON Web Token) ##
  #
  # The [`id_token` mutator](https://www.ory.sh/oathkeeper/docs/pipeline/mutator#id_token).
  #
  id_token:
    
    ## config ##
    #
    config:
      
      ## jwks_url ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export MUTATORS_ID_TOKEN_CONFIG_JWKS_URL=<value>
      # - Windows Command Line (CMD):
      #    > set MUTATORS_ID_TOKEN_CONFIG_JWKS_URL=<value>
      #
      jwks_url: https://fetch-keys/from/this/location.json

      ## issuer_url ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export MUTATORS_ID_TOKEN_CONFIG_ISSUER_URL=<value>
      # - Windows Command Line (CMD):
      #    > set MUTATORS_ID_TOKEN_CONFIG_ISSUER_URL=<value>
      #
      issuer_url: eiusmod ut in

      ## claims ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export MUTATORS_ID_TOKEN_CONFIG_CLAIMS=<value>
      # - Windows Command Line (CMD):
      #    > set MUTATORS_ID_TOKEN_CONFIG_CLAIMS=<value>
      #
      claims: tempor reprehenderit

      ## ttl ##
      #
      # Set this value using environment variables on
      # - Linux/macOS:
      #    $ export MUTATORS_ID_TOKEN_CONFIG_TTL=<value>
      # - Windows Command Line (CMD):
      #    > set MUTATORS_ID_TOKEN_CONFIG_TTL=<value>
      #
      ttl: 1h

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
    enabled: true

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
  # One of:
  # - panic
  # - fatal
  # - error
  # - warn
  # - info
  # - debug
  # 
  # Set this value using environment variables on
  # - Linux/macOS:
  #    $ export LOG_LEVEL=<value>
  # - Windows Command Line (CMD):
  #    > set LOG_LEVEL=<value>
  #
  level: warn

  ## Format ##
  #
  # The log format can either be text or JSON.
  #
  # Default value: text
  #
  # One of:
  # - text
  # - json
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
# One of:
# - cpu
# - mem
# - ""
# 
# Set this value using environment variables on
# - Linux/macOS:
#    $ export PROFILING=<value>
# - Windows Command Line (CMD):
#    > set PROFILING=<value>
#
profiling: mem

## The Oathkeeper version this config is written for. ##
#
# SemVer according to https://semver.org/ prefixed with `v` as in our releases.
#
# Set this value using environment variables on
# - Linux/macOS:
#    $ export VERSION=<value>
# - Windows Command Line (CMD):
#    > set VERSION=<value>
#
version: v0.0.2199-3009880765dC.0.0+0QRkqQ9.Pk.dVjFK4HJW.4.DQexyvAS.9.zFr.QNn

```