package configuration

type Key = string

const (
	ViperKeyProxyReadTimeout                    Key = "serve.proxy.timeout.read"
	ViperKeyProxyWriteTimeout                   Key = "serve.proxy.timeout.write"
	ViperKeyProxyIdleTimeout                    Key = "serve.proxy.timeout.idle"
	ViperKeyProxyServeAddressHost               Key = "serve.proxy.host"
	ViperKeyProxyServeAddressPort               Key = "serve.proxy.port"
	ViperKeyAPIServeAddressHost                 Key = "serve.api.host"
	ViperKeyAPIServeAddressPort                 Key = "serve.api.port"
	ViperKeyAPIReadTimeout                      Key = "serve.api.timeout.read"
	ViperKeyAPIWriteTimeout                     Key = "serve.api.timeout.write"
	ViperKeyAPIIdleTimeout                      Key = "serve.api.timeout.idle"
	ViperKeyPrometheusServeAddressHost          Key = "serve.prometheus.host"
	ViperKeyPrometheusServeAddressPort          Key = "serve.prometheus.port"
	ViperKeyPrometheusServeMetricsPath          Key = "serve.prometheus.metrics_path"
	ViperKeyPrometheusServeMetricsNamePrefix    Key = "serve.prometheus.metric_name_prefix"
	ViperKeyPrometheusServeHideRequestPaths     Key = "serve.prometheus.hide_request_paths"
	ViperKeyPrometheusServeCollapseRequestPaths Key = "serve.prometheus.collapse_request_paths"
	ViperKeyAccessRuleRepositories              Key = "access_rules.repositories"
	ViperKeyAccessRuleMatchingStrategy          Key = "access_rules.matching_strategy"
)

// Authorizers
const (
	ViperKeyAuthorizerAllowIsEnabled            Key = "authorizers.allow.enabled"
	ViperKeyAuthorizerDenyIsEnabled             Key = "authorizers.deny.enabled"
	ViperKeyAuthorizerKetoEngineACPORYIsEnabled Key = "authorizers.keto_engine_acp_ory.enabled"
	ViperKeyAuthorizerRemoteIsEnabled           Key = "authorizers.remote.enabled"
	ViperKeyAuthorizerRemoteJSONIsEnabled       Key = "authorizers.remote_json.enabled"
)

// Mutators
const (
	ViperKeyMutatorCookieIsEnabled   Key = "mutators.cookie.enabled"
	ViperKeyMutatorHeaderIsEnabled   Key = "mutators.header.enabled"
	ViperKeyMutatorNoopIsEnabled     Key = "mutators.noop.enabled"
	ViperKeyMutatorHydratorIsEnabled Key = "mutators.hydrator.enabled"
	ViperKeyMutatorIDTokenIsEnabled  Key = "mutators.id_token.enabled"
	ViperKeyMutatorIDTokenJWKSURL    Key = "mutators.id_token.config.jwks_url"
)

// Authenticators
const (
	// anonymous
	ViperKeyAuthenticatorAnonymousIsEnabled Key = "authenticators.anonymous.enabled"

	// noop
	ViperKeyAuthenticatorNoopIsEnabled Key = "authenticators.noop.enabled"

	// cookie session
	ViperKeyAuthenticatorCookieSessionIsEnabled Key = "authenticators.cookie_session.enabled"

	// jwt
	ViperKeyAuthenticatorJwtIsEnabled  Key = "authenticators.jwt.enabled"
	ViperKeyAuthenticatorJwtJwkMaxWait Key = "authenticators.jwt.config.jwks_max_wait"
	ViperKeyAuthenticatorJwtJwkTtl     Key = "authenticators.jwt.config.jwks_ttl"

	// oauth2_client_credentials
	ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled Key = "authenticators.oauth2_client_credentials.enabled"

	// oauth2_token_introspection
	ViperKeyAuthenticatorOAuth2TokenIntrospectionIsEnabled Key = "authenticators.oauth2_introspection.enabled"

	// unauthorized
	ViperKeyAuthenticatorUnauthorizedIsEnabled Key = "authenticators.unauthorized.enabled"
)

// Errors
const (
	ViperKeyErrors                         Key = "errors.handlers"
	ViperKeyErrorsFallback                 Key = "errors.fallback"
	ViperKeyErrorsJSONIsEnabled            Key = ViperKeyErrors + ".json.enabled"
	ViperKeyErrorsRedirectIsEnabled        Key = ViperKeyErrors + ".redirect.enabled"
	ViperKeyErrorsWWWAuthenticateIsEnabled Key = ViperKeyErrors + ".www_authenticate.enabled"
)
