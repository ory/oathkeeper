package configuration

import (
	"github.com/ory/fosite"
	"github.com/ory/x/viperx"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
	"net/url"
	"strings"
)

const (
	ViperKeyAuthenticatorAnonymousEnabled  = "authenticators.anonymous.enabled"
	ViperKeyAuthenticatorAnonymousUsername = "authenticators.anonymous.username"

	ViperKeyAuthenticatorNoopEnabled = "authenticators.noop.enabled"

	ViperKeyAuthenticatorJWTEnabled       = "authenticators.jwt.enabled"
	ViperKeyAuthenticatorJWTJWKSURIs      = "authenticators.jwt.jwk_urls"
	ViperKeyAuthenticatorJWTScopeStrategy = "authenticators.jwt.scope_strategy"

	ViperKeyAuthenticatorClientCredentialsTokenURL = "authenticators.oauth2_client_credentials.token_url"

	ViperKeyAuthenticatorOAuth2TokenIntrospectionScopeStrategy    = "authenticators.oauth2_introspection.scope_strategy"
	ViperKeyAuthenticatorOAuth2TokenIntrospectionIntrospectionURL = "authenticators.oauth2_introspection.introspection_url"
)

type ViperProvider struct {
	l logrus.FieldLogger
}

func (v *ViperProvider) AuthenticatorAnonymousEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyAuthenticatorAnonymousEnabled, true)

}

func (v *ViperProvider) AuthenticatorAnonymousIdentifier() string {
	return viperx.GetString(v.l, ViperKeyAuthenticatorAnonymousUsername, "anonymous", "AUTHENTICATOR_ANONYMOUS_USERNAME")
}

func (v *ViperProvider) AuthenticatorNoopEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyAuthenticatorNoopEnabled, false)
}

func (v *ViperProvider) AuthenticatorJWTEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyAuthenticatorJWTEnabled, false)
}

func (v *ViperProvider) AuthenticatorJWTJWKSURIs() []string {
	return viperx.GetStringSlice(v.l, ViperKeyAuthenticatorJWTJWKSURIs, []string{}, "AUTHENTICATOR_JWT_JWKS_URL")
}

func (v *ViperProvider) AuthenticatorJWTScopeStrategy() fosite.ScopeStrategy {
	return v.toScopeStrategy(
		viperx.GetString(v.l, ViperKeyAuthenticatorJWTScopeStrategy, "none", "AUTHENTICATOR_JWT_SCOPE_STRATEGY"),
		ViperKeyAuthenticatorJWTScopeStrategy,
	)
}

func (v *ViperProvider) AuthenticatorClientCredentialsTokenURL() *url.URL {
	return v.getURL(
		viperx.GetString(v.l, ViperKeyAuthenticatorClientCredentialsTokenURL, "", "AUTHENTICATOR_OAUTH2_CLIENT_CREDENTIALS_TOKEN_URL"),
		ViperKeyAuthenticatorClientCredentialsTokenURL,
	)
}

func (v *ViperProvider) AuthenticatorOAuth2TokenIntrospectionScopeStrategy() fosite.ScopeStrategy {
	return v.toScopeStrategy(
		viperx.GetString(v.l, ViperKeyAuthenticatorOAuth2TokenIntrospectionScopeStrategy, "none", "AUTHENTICATOR_OAUTH2_INTROSPECTION_SCOPE_STRATEGY"),
		ViperKeyAuthenticatorOAuth2TokenIntrospectionScopeStrategy,
	)
}

func (v *ViperProvider) AuthenticatorOAuth2TokenIntrospectionIntrospectionURL() *url.URL {
	return v.getURL(
		viperx.GetString(v.l, ViperKeyAuthenticatorOAuth2TokenIntrospectionIntrospectionURL, "", "AUTHENTICATOR_OAUTH2_INTROSPECTION_URL"),
		ViperKeyAuthenticatorOAuth2TokenIntrospectionIntrospectionURL,
	)
}

func (v *ViperProvider) AuthenticatorOAuth2TokenIntrospectionPreAuthorization() *clientcredentials.Config {

}

func (v *ViperProvider) getURL(value string, key string) *url.URL {
	u, err := url.ParseRequestURI(value)
	if err != nil {
		v.l.WithError(err).Errorf(`Configuration key "%s" is missing or malformed.`, key)
		return nil
	}

	return u
}

func (v *ViperProvider) toScopeStrategy(value string, key string) fosite.ScopeStrategy {
	switch strings.ToLower(value) {
	case "hierarchic":
		return fosite.HierarchicScopeStrategy
	case "exact":
		return fosite.ExactScopeStrategy
	case "wildcard":
		return fosite.WildcardScopeStrategy
	case "none":
		return nil
	default:
		v.l.Errorf(`Configuration key "%s" declares unknown scope strategy "%s", only "hierarchic", "exact", "wildcard", "none" are supported. Falling back to strategy "none".`, key, value)
		return nil
	}
}
