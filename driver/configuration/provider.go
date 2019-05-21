package configuration

import (
	"net/url"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/rs/cors"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/fosite"
)

type Provider interface {
	CORSEnabled(iface string) bool
	CORSOptions(iface string) cors.Options

	ProviderAuthenticators
	ProviderAuthorizers
	ProviderMutators

	ProxyReadTimeout() time.Duration
	ProxyWriteTimeout() time.Duration
	ProxyIdleTimeout() time.Duration

	AccessRuleRepositories() []url.URL

	ProxyServeAddress() string
	APIServeAddress() string
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

	AuthenticatorHiveIsEnabled() bool
	AuthenticatorHivePublicURL() *url.URL
	AuthenticatorHiveAdminURL() *url.URL
}

type ProviderAuthorizers interface {
	AuthorizerAllowIsEnabled() bool

	AuthorizerDenyIsEnabled() bool

	AuthorizerKetoEngineACPORYIsEnabled() bool
	AuthorizerKetoEngineACPORYBaseURL() *url.URL
}

type ProviderMutators interface {
	MutatorCookieIsEnabled() bool

	MutatorHeaderIsEnabled() bool

	MutatorIDTokenIsEnabled() bool
	MutatorIDTokenIssuerURL() *url.URL
	MutatorIDTokenJWKSURL() *url.URL
	MutatorIDTokenTTL() time.Duration

	MutatorNoopIsEnabled() bool
}

func MustValidate(l logrus.FieldLogger, p Provider) {
}
