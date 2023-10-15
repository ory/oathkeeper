// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package configuration

type Key = string

const (
	ProxyReadTimeout                    Key = "serve.proxy.timeout.read"
	ProxyWriteTimeout                   Key = "serve.proxy.timeout.write"
	ProxyIdleTimeout                    Key = "serve.proxy.timeout.idle"
	ProxyServeAddressHost               Key = "serve.proxy.host"
	ProxyServeAddressPort               Key = "serve.proxy.port"
	ProxyTrustForwardedHeaders          Key = "serve.proxy.trust_forwarded_headers"
	APIServeAddressHost                 Key = "serve.api.host"
	APIServeAddressPort                 Key = "serve.api.port"
	APIReadTimeout                      Key = "serve.api.timeout.read"
	APIWriteTimeout                     Key = "serve.api.timeout.write"
	APIIdleTimeout                      Key = "serve.api.timeout.idle"
	PrometheusServeAddressHost          Key = "serve.prometheus.host"
	PrometheusServeAddressPort          Key = "serve.prometheus.port"
	PrometheusServeMetricsPath          Key = "serve.prometheus.metrics_path"
	PrometheusServeMetricsNamePrefix    Key = "serve.prometheus.metric_name_prefix"
	PrometheusServeHideRequestPaths     Key = "serve.prometheus.hide_request_paths"
	PrometheusServeCollapseRequestPaths Key = "serve.prometheus.collapse_request_paths"
	AccessRuleRepositories              Key = "access_rules.repositories"
	AccessRuleMatchingStrategy          Key = "access_rules.matching_strategy"
)

// Authorizers
const (
	AuthorizerAllowIsEnabled            Key = "authorizers.allow.enabled"
	AuthorizerDenyIsEnabled             Key = "authorizers.deny.enabled"
	AuthorizerKetoEngineACPORYIsEnabled Key = "authorizers.keto_engine_acp_ory.enabled"
	AuthorizerRemoteIsEnabled           Key = "authorizers.remote.enabled"
	AuthorizerRemoteJSONIsEnabled       Key = "authorizers.remote_json.enabled"
)

// Mutators
const (
	MutatorCookieIsEnabled   Key = "mutators.cookie.enabled"
	MutatorHeaderIsEnabled   Key = "mutators.header.enabled"
	MutatorNoopIsEnabled     Key = "mutators.noop.enabled"
	MutatorHydratorIsEnabled Key = "mutators.hydrator.enabled"
	MutatorIDTokenIsEnabled  Key = "mutators.id_token.enabled"
	MutatorIDTokenJWKSURL    Key = "mutators.id_token.config.jwks_url"
	MutatorIDTokenIssuerURL  Key = "mutators.id_token.config.issuer_url"
)

// Authenticators
const (
	// anonymous
	AuthenticatorAnonymousIsEnabled Key = "authenticators.anonymous.enabled"

	// noop
	AuthenticatorNoopIsEnabled Key = "authenticators.noop.enabled"

	// cookie session
	AuthenticatorCookieSessionIsEnabled Key = "authenticators.cookie_session.enabled"

	// jwt
	AuthenticatorJwtIsEnabled  Key = "authenticators.jwt.enabled"
	AuthenticatorJwtJwkMaxWait Key = "authenticators.jwt.config.jwks_max_wait"
	AuthenticatorJwtJwkTtl     Key = "authenticators.jwt.config.jwks_ttl"

	// oauth2_client_credentials
	AuthenticatorOAuth2ClientCredentialsIsEnabled Key = "authenticators.oauth2_client_credentials.enabled"

	// oauth2_token_introspection
	AuthenticatorOAuth2TokenIntrospectionIsEnabled Key = "authenticators.oauth2_introspection.enabled"

	// unauthorized
	AuthenticatorUnauthorizedIsEnabled Key = "authenticators.unauthorized.enabled"
)

// Errors
const (
	ErrorsHandlers                 Key = "errors.handlers"
	ErrorsFallback                 Key = "errors.fallback"
	ErrorsJSONIsEnabled            Key = ErrorsHandlers + ".json.enabled"
	ErrorsRedirectIsEnabled        Key = ErrorsHandlers + ".redirect.enabled"
	ErrorsWWWAuthenticateIsEnabled Key = ErrorsHandlers + ".www_authenticate.enabled"
)
