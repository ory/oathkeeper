package configuration

import (
	"github.com/ory/fosite"
	"github.com/rs/cors"
	"golang.org/x/oauth2/clientcredentials"
	"net/url"
)

type Provider interface {
	ServesHTTPS() bool

	CORSEnabled(iface string) bool
	CORSOptions(iface string) cors.Options

	AuthenticatorAnonymousIsEnabled() bool
	AuthenticatorAnonymousIdentifier() string

	AuthenticatorNoopIsEnabled() bool

	AuthenticatorJWTIsEnabled() bool
	AuthenticatorJWTJWKSURIs() []url.URL
	AuthenticatorJWTScopeStrategy() fosite.ScopeStrategy

	AuthenticatorOAuth2ClientCredentialsIsEnabled() bool
	AuthenticatorOAuth2ClientCredentialsTokenURL() *url.URL

	AuthenticatorOAuth2TokenIntrospectionScopeStrategy() fosite.ScopeStrategy
	AuthenticatorOAuth2TokenIntrospectionIntrospectionURL() *url.URL
	AuthenticatorOAuth2TokenIntrospectionAuthorization() *clientcredentials.Config

	AuthorizerAllowIsEnabled() bool
	AuthorizerDenyIsEnabled() bool

	AuthorizerKetoWardenIsEnabled() bool
}
