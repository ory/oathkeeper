package authn

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/ory/go-convenience/stringslice"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/x/httpx"
)

type AuthenticatorOAuth2IntrospectionConfiguration struct {
	// An array of OAuth 2.0 scopes that are required when accessing an endpoint protected by this handler.
	// If the token used in the Authorization header did not request that specific scope, the request is denied.
	Scopes []string `json:"required_scope"`

	// An array of audiences that are required when accessing an endpoint protected by this handler.
	// If the token used in the Authorization header is not intended for any of the requested audiences, the request is denied.
	Audience []string `json:"target_audience"`

	// The token must have been issued by one of the issuers listed in this array.
	Issuers []string `json:"trusted_issuers"`
}

type AuthenticatorOAuth2Introspection struct {
	c configuration.Provider

	client *http.Client
}

func NewAuthenticatorOAuth2Introspection(c configuration.Provider) *AuthenticatorOAuth2Introspection {
	var rt http.RoundTripper
	if conf := c.AuthenticatorOAuth2TokenIntrospectionPreAuthorization(); conf != nil {
		rt = conf.Client(context.Background()).Transport
	}

	return &AuthenticatorOAuth2Introspection{c: c, client: httpx.NewResilientClientLatencyToleranceSmall(rt)}
}

func (a *AuthenticatorOAuth2Introspection) GetID() string {
	return "oauth2_introspection"
}

type AuthenticatorOAuth2IntrospectionResult struct {
	Active    bool                   `json:"active"`
	Extra     map[string]interface{} `json:"ext"`
	Subject   string                 `json:"sub,omitempty"`
	Username  string                 `json:"username"`
	Audience  []string               `json:"aud"`
	TokenType string                 `json:"token_type"`
	Issuer    string                 `json:"iss"`
	ClientID  string                 `json:"client_id,omitempty"`
	Scope     string                 `json:"scope,omitempty"`
}

func (a *AuthenticatorOAuth2Introspection) Authenticate(r *http.Request, config json.RawMessage, _ pipeline.Rule) (*AuthenticationSession, error) {
	var i AuthenticatorOAuth2IntrospectionResult
	var cf AuthenticatorOAuth2IntrospectionConfiguration

	if len(config) == 0 {
		config = []byte("{}")
	}

	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cf); err != nil {
		return nil, errors.WithStack(err)
	}

	token := helper.BearerTokenFromRequest(r)
	if token == "" {
		return nil, errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	body := url.Values{"token": {token}, "scope": {strings.Join(cf.Scopes, " ")}}
	resp, err := a.client.Post(a.c.AuthenticatorOAuth2TokenIntrospectionIntrospectionURL().String(), "application/x-www-form-urlencoded", strings.NewReader(body.Encode()))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Introspection returned status code %d but expected %d", resp.StatusCode, http.StatusOK)
	}

	if err := json.NewDecoder(resp.Body).Decode(&i); err != nil {
		return nil, errors.WithStack(err)
	}

	if len(i.TokenType) > 0 && i.TokenType != "access_token" {
		return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Introspected token is not an access token but \"%s\"", i.TokenType)))
	}

	if !i.Active {
		return nil, errors.WithStack(helper.ErrForbidden.WithReason("Access token i says token is not active"))
	}

	for _, audience := range cf.Audience {
		if !stringslice.Has(i.Audience, audience) {
			return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Token audience is not intended for target audience %s", audience)))
		}
	}

	if len(cf.Issuers) > 0 {
		if !stringslice.Has(cf.Issuers, i.Issuer) {
			return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Token issuer does not match any trusted issuer")))
		}
	}

	if a.c.AuthenticatorOAuth2TokenIntrospectionScopeStrategy() != nil {
		for _, scope := range cf.Scopes {
			if !a.c.AuthenticatorOAuth2TokenIntrospectionScopeStrategy()(strings.Split(i.Scope, " "), scope) {
				return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Scope %s was not granted", scope)))
			}
		}
	}

	if len(i.Extra) == 0 {
		i.Extra = map[string]interface{}{}
	}

	i.Extra["username"] = i.Username
	i.Extra["client_id"] = i.ClientID
	i.Extra["scope"] = i.Scope

	return &AuthenticationSession{
		Subject: i.Subject,
		Extra:   i.Extra,
	}, nil
}

func (a *AuthenticatorOAuth2Introspection) Validate() error {
	if !a.c.AuthenticatorOAuth2TokenIntrospectionIsEnabled() {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf(`Authenticator "%s" is disabled per configuration.`, a.GetID()))
	}

	if a.c.AuthenticatorOAuth2TokenIntrospectionIntrospectionURL() == nil {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf(`Configuration for authenticator "%s" did not specify any values for configuration key "%s" and is thus disabled.`, a.GetID(), configuration.ViperKeyAuthenticatorOAuth2TokenIntrospectionIntrospectionURL))
	}

	return nil
}
