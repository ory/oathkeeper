package configuration

import (
	"net/url"

	"github.com/ory/x/viperx"
)

const (
	ViperKeyAuthorizerAllowIsEnabled = "authorizers.allow.enabled"

	ViperKeyAuthorizerDenyIsEnabled = "authorizers.deny.enabled"

	ViperKeyAuthorizerKetoEngineACPORYIsEnabled = "authorizers.keto_engine_acp_ory.enabled"
	ViperKeyAuthorizerKetoEngineACPORYBaseURL   = "authorizers.keto_engine_acp_ory.base_url"

	ViperKeyAuthorizerOPAEngineIsEnabled = "authorizers.opa_engine.enabled"
	ViperKeyAuthorizerOPAEngineBaseURL   = "authorizers.opa_engine.base_url"
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

func (v *ViperProvider) AuthorizerKetoEngineACPORYBaseURL() *url.URL {
	return v.getURL(
		viperx.GetString(v.l, ViperKeyAuthorizerKetoEngineACPORYBaseURL, "", "AUTHORIZER_KETO_URL"),
		ViperKeyAuthorizerKetoEngineACPORYBaseURL,
	)
}

func (v *ViperProvider) AuthorizerOPAEngineIsEnabled() bool {
	return viperx.GetBool(v.l, ViperKeyAuthorizerOPAEngineIsEnabled, false)
}

func (v *ViperProvider) AuthorizerOPAEngineBaseURL() *url.URL {
	return v.getURL(
		viperx.GetString(v.l, ViperKeyAuthorizerOPAEngineBaseURL, "", "AUTHORIZER_OPA_URL"),
		ViperKeyAuthorizerOPAEngineBaseURL,
	)
}
