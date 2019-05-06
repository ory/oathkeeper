package configuration

import (
	"github.com/ory/fosite"
	"github.com/rs/cors"
	"golang.org/x/oauth2/clientcredentials"
	"net/url"
	"time"
)

type Provider interface {
	ServesHTTPS() bool

	CORSEnabled(iface string) bool
	CORSOptions(iface string) cors.Options

	ProviderAuthenticators
	ProviderAuthorizers
	ProviderTransformers
}

type ProviderAuthenticators interface {
	AuthenticatorAnonymousIsEnabled() bool
	AuthenticatorAnonymousIdentifier() string

	AuthenticatorNoopIsEnabled() bool

	AuthenticatorJWTIsEnabled() bool
	AuthenticatorJWTJWKSURIs() []url.URL
	AuthenticatorJWTScopeStrategy() fosite.ScopeStrategy

	AuthenticatorOAuth2ClientCredentialsIsEnabled() bool
	AuthenticatorOAuth2ClientCredentialsTokenURL() *url.URL

	AuthenticatorOAuth2TokenIntrospectionIsEnabled() bool
	AuthenticatorOAuth2TokenIntrospectionScopeStrategy() fosite.ScopeStrategy
	AuthenticatorOAuth2TokenIntrospectionIntrospectionURL() *url.URL
	AuthenticatorOAuth2TokenIntrospectionPreAuthorization() *clientcredentials.Config

	AuthenticatorUnauthorizedIsEnabled() bool
}

type ProviderAuthorizers interface {
	AuthorizerAllowIsEnabled() bool

	AuthorizerDenyIsEnabled() bool

	AuthorizerKetoEngineACPORYIsEnabled() bool
	AuthorizerKetoEngineACPORYAuthorizedURL() *url.URL
}

type ProviderTransformers interface {
	TransformerCookieIsEnabled() bool

	TransformerHeaderIsEnabled() bool

	TransformerIDTokenIsEnabled() bool
	TransformerIDTokenIssuerURL() *url.URL
	TransformerIDTokenTTL() time.Duration

	TransformerNoopIsEnabled() bool
}
