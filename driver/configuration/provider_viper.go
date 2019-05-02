package configuration

import (
	"github.com/ory/fosite"
	"github.com/ory/x/viperx"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	ViperKeyAuthenticatorAnonymousEnabled  = "authenticators.anonymous.enabled"
	ViperKeyAuthenticatorAnonymousUsername = "authenticators.anonymous.username"

	ViperKeyAuthenticatorNoopEnabled = "authenticators.noop.enabled"

	ViperKeyAuthenticatorJWTEnabled       = "authenticators.jwt.enabled"
	ViperKeyAuthenticatorJWTJWKSURIs      = "authenticators.jwt.jwk_uris"
	ViperKeyAuthenticatorJWTScopeStrategy = "authenticators.jwt.scope_strategy"
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
	switch id := viperx.GetString(v.l, ViperKeyAuthenticatorJWTScopeStrategy, "exact", "AUTHENTICATOR_JWT_SCOPE_STRATEGY"); strings.ToLower(id) {
	case "hierarchic":
		return fosite.HierarchicScopeStrategy
	case "exact":
		return fosite.ExactScopeStrategy
	case "wildcard":
		return fosite.WildcardScopeStrategy
	case "none":
		fallthrough
	default:
		v.l.Warnf(`Configuration key "%s" declares unknown scope strategy "%s", only "hierarchic", "exact", "wildcard", "none" are supported. Falling back to strategy "none".`, ViperKeyAuthenticatorJWTScopeStrategy, id)
		return nil
	}
	return nil
}
