package configuration

import (
	"net/url"

	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/fosite"
	"github.com/ory/x/viperx"
)

// Authenticators
const (
	// anonymous
	ViperKeyAuthenticatorAnonymousIsEnabled  = "authenticators.anonymous.enabled"
	ViperKeyAuthenticatorAnonymousIdentifier = "authenticators.anonymous.subject"

	// noop
	ViperKeyAuthenticatorNoopIsEnabled = "authenticators.noop.enabled"

	// jwt
	ViperKeyAuthenticatorJWTIsEnabled     = "authenticators.jwt.enabled"
	ViperKeyAuthenticatorJWTJWKSURIs      = "authenticators.jwt.jwks_urls"
	ViperKeyAuthenticatorJWTScopeStrategy = "authenticators.jwt.scope_strategy"

	// oauth2_client_credentials
	ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled = "authenticators.oauth2_client_credentials.enabled"
	ViperKeyAuthenticatorClientCredentialsTokenURL        = "authenticators.oauth2_client_credentials.token_url"

	// oauth2_token_introspection
	ViperKeyAuthenticatorOAuth2TokenIntrospectionIsEnabled                    = "authenticators.oauth2_introspection.enabled"
	ViperKeyAuthenticatorOAuth2TokenIntrospectionScopeStrategy                = "authenticators.oauth2_introspection.scope_strategy"
	ViperKeyAuthenticatorOAuth2TokenIntrospectionIntrospectionURL             = "authenticators.oauth2_introspection.introspection_url"
	ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationEnabled      = "authenticators.oauth2_introspection.pre_authorization.enabled"
	ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationClientID     = "authenticators.oauth2_introspection.pre_authorization.client_id"
	ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationClientSecret = "authenticators.oauth2_introspection.pre_authorization.client_secret"
	ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationScope        = "authenticators.oauth2_introspection.pre_authorization.scope"
	ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationTokenURL     = "authenticators.oauth2_introspection.pre_authorization.token_url"

	// unauthorized
	ViperKeyAuthenticatorUnauthorizedIsEnabled = "authenticators.unauthorized.enabled"
)

func (v *ViperProvider) AuthenticatorAnonymousIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyAuthenticatorAnonymousIsEnabled, false)
}
func (v *ViperProvider) AuthenticatorAnonymousIdentifier() string {
	return viperx.GetString(v.l, ViperKeyAuthenticatorAnonymousIdentifier, "anonymous", "AUTHENTICATOR_ANONYMOUS_USERNAME")
}

func (v *ViperProvider) AuthenticatorNoopIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyAuthenticatorNoopIsEnabled, false)

}

func (v *ViperProvider) AuthenticatorJWTIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyAuthenticatorJWTIsEnabled, false)

}
func (v *ViperProvider) AuthenticatorJWTJWKSURIs() []url.URL {
	res := make([]url.URL, 0)
	for _, u := range viperx.GetStringSlice(v.l, ViperKeyAuthenticatorJWTJWKSURIs, []string{}, "AUTHENTICATOR_JWT_JWKS_URL") {
		if p := v.getURL(u, ViperKeyAuthenticatorJWTJWKSURIs); p != nil {
			res = append(res, *p)
		}
	}
	return res
}

func (v *ViperProvider) AuthenticatorJWTScopeStrategy() fosite.ScopeStrategy {
	return v.toScopeStrategy(
		viperx.GetString(v.l, ViperKeyAuthenticatorJWTScopeStrategy, "none", "AUTHENTICATOR_JWT_SCOPE_STRATEGY"),
		ViperKeyAuthenticatorJWTScopeStrategy,
	)
}

func (v *ViperProvider) AuthenticatorOAuth2ClientCredentialsIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyAuthenticatorOAuth2ClientCredentialsIsEnabled, false)

}
func (v *ViperProvider) AuthenticatorOAuth2ClientCredentialsTokenURL() *url.URL {
	return v.getURL(
		viperx.GetString(v.l, ViperKeyAuthenticatorClientCredentialsTokenURL, "", "AUTHENTICATOR_OAUTH2_CLIENT_CREDENTIALS_TOKEN_URL"),
		ViperKeyAuthenticatorClientCredentialsTokenURL,
	)
}

func (v *ViperProvider) AuthenticatorOAuth2TokenIntrospectionIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyAuthenticatorOAuth2TokenIntrospectionIsEnabled, false)
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
	if !viperx.GetBool(v.l, ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationEnabled, false) {
		v.l.Infof("Authenticator oauth2_token_introspection did not specify pre-authorization which is thus disabled")
		return nil
	}

	var (
		id     = viperx.GetString(v.l, ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationClientID, "", "AUTHENTICATOR_OAUTH2_INTROSPECTION_AUTHORIZATION_CLIENT_ID")
		secret = viperx.GetString(v.l, ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationClientSecret, "", "AUTHENTICATOR_OAUTH2_INTROSPECTION_AUTHORIZATION_CLIENT_SECRET")
		tu     = viperx.GetString(v.l, ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationTokenURL, "", "AUTHENTICATOR_OAUTH2_INTROSPECTION_AUTHORIZATION_TOKEN_URL")
		scope  = viperx.GetStringSlice(v.l, ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationScope, []string{}, "AUTHENTICATOR_OAUTH2_INTROSPECTION_AUTHORIZATION_SCOPE")
	)

	if len(id) == 0 {
		v.l.Errorf(`Authenticator oauth2_token_introspection has pre-authorization enabled but configuration value "%s" is missing or empty. Thus, pre-authorization is disabled.`, ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationClientID)
		return nil
	}

	if len(secret) == 0 {
		v.l.Errorf(`Authenticator oauth2_token_introspection has pre-authorization enabled but configuration value "%s" is missing or empty. Thus, pre-authorization is disabled.`, ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationClientSecret)
		return nil
	}

	if len(tu) == 0 {
		v.l.Errorf(`Authenticator oauth2_token_introspection has pre-authorization enabled but configuration value "%s" is missing or empty. Thus, pre-authorization is disabled.`, ViperKeyAuthenticatorOAuth2TokenIntrospectionPreAuthorizationTokenURL)
		return nil
	}

	return &clientcredentials.Config{
		ClientID:     id,
		ClientSecret: secret,
		Scopes:       scope,
		TokenURL:     tu,
	}
}

func (v *ViperProvider) AuthenticatorUnauthorizedIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyAuthenticatorUnauthorizedIsEnabled, false)
}
