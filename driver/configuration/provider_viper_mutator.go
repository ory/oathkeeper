package configuration

import (
	"net/url"
	"time"

	"github.com/ory/x/viperx"
)

const (
	ViperKeyMutatorCookieIsEnabled = "mutators.cookie.enabled"

	ViperKeyMutatorHeaderIsEnabled = "mutators.header.enabled"

	ViperKeyMutatorNoopIsEnabled = "mutators.noop.enabled"

	ViperKeyMutatorIDTokenIsEnabled = "mutators.id_token.enabled"
	ViperKeyMutatorIDTokenIssuerURL = "mutators.id_token.issuer_url"
	ViperKeyMutatorIDTokenJWKSURL   = "mutators.id_token.jwks_url"
	ViperKeyMutatorIDTokenTTL       = "mutators.id_token.ttl"
)

func (v *ViperProvider) MutatorCookieIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyMutatorCookieIsEnabled, false)
}

func (v *ViperProvider) MutatorHeaderIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyMutatorHeaderIsEnabled, false)

}

func (v *ViperProvider) MutatorIDTokenIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyMutatorIDTokenIsEnabled, false)
}
func (v *ViperProvider) MutatorIDTokenIssuerURL() *url.URL {
	return v.getURL(
		viperx.GetString(v.l, ViperKeyMutatorIDTokenIssuerURL, "", "CREDENTIALS_ISSUER_ID_TOKEN_ISSUER"),
		ViperKeyMutatorIDTokenIssuerURL,
	)
}
func (v *ViperProvider) MutatorIDTokenJWKSURL() *url.URL {
	return v.getURL(
		viperx.GetString(v.l, ViperKeyMutatorIDTokenJWKSURL, ""),
		ViperKeyMutatorIDTokenJWKSURL,
	)
}
func (v *ViperProvider) MutatorIDTokenTTL() time.Duration {
	return viperx.GetDuration(v.l, ViperKeyMutatorIDTokenTTL, time.Minute, "CREDENTIALS_ISSUER_ID_TOKEN_LIFESPAN")
}

func (v *ViperProvider) MutatorNoopIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyMutatorNoopIsEnabled, false)
}
