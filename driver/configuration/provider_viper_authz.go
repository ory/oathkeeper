package configuration

import (
	"net/url"

	"github.com/ory/x/viperx"
)

const (
	ViperKeyAuthorizerAllowIsEnabled = "authorizers.allow.enabled"

	ViperKeyAuthorizerDenyIsEnabled = "authorizers.deny.enabled"

	ViperKeyAuthorizerKetoEngineACPORYIsEnabled     = "authorizers.keto_engine_acp_ory.enabled"
	ViperKeyAuthorizerKetoEngineACPORYAuthorizedURL = "authorizers.keto_engine_acp_ory.base_url"
)

func (v *ViperProvider) AuthorizerAllowIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyAuthorizerAllowIsEnabled, false)
}

func (v *ViperProvider) AuthorizerDenyIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyAuthorizerDenyIsEnabled, false)
}

func (v *ViperProvider) AuthorizerKetoEngineACPORYIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyAuthorizerKetoEngineACPORYIsEnabled, false)
}

func (v *ViperProvider) AuthorizerKetoEngineACPORYAuthorizedURL() *url.URL {
	return v.getURL(
		viperx.GetString(v.l, ViperKeyAuthorizerKetoEngineACPORYAuthorizedURL, "", "AUTHORIZER_KETO_URL"),
		ViperKeyAuthorizerKetoEngineACPORYAuthorizedURL,
	)
}
