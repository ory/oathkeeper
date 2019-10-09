package authn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/go-convenience/stringslice"
	"github.com/ory/x/httpx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
)

type AuthenticatorOAuth2IntrospectionConfiguration struct {
	Scopes              []string                                              `json:"required_scope"`
	Audience            []string                                              `json:"target_audience"`
	Issuers             []string                                              `json:"trusted_issuers"`
	PreAuth             *AuthenticatorOAuth2IntrospectionPreAuthConfiguration `json:"pre_authorization"`
	ScopeStrategy       string                                                `json:"scope_strategy"`
	IntrospectionURL    string                                                `json:"introspection_url"`
	BearerTokenLocation *helper.BearerTokenLocation                           `json:"token_from"`
}

type AuthenticatorOAuth2IntrospectionPreAuthConfiguration struct {
	Enabled      bool     `json:"enabled"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Scope        []string `json:"scope"`
	TokenURL     string   `json:"token_url"`
}

type AuthenticatorOAuth2Introspection struct {
	c configuration.Provider

	client *http.Client
}

func NewAuthenticatorOAuth2Introspection(c configuration.Provider) *AuthenticatorOAuth2Introspection {
	var rt http.RoundTripper
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
	cf, err := a.Config(config)
	if err != nil {
		return nil, err
	}

	token := helper.BearerTokenFromRequest(r, cf.BearerTokenLocation)
	if token == "" {
		return nil, errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	body := url.Values{"token": {token}, "scope": {strings.Join(cf.Scopes, " ")}}
	resp, err := a.client.Post(cf.IntrospectionURL, "application/x-www-form-urlencoded", strings.NewReader(body.Encode()))
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

	if ss := a.c.ToScopeStrategy(cf.ScopeStrategy, "authenticators.oauth2_introspection.scope_strategy"); ss != nil {
		for _, scope := range cf.Scopes {
			if !ss(strings.Split(i.Scope, " "), scope) {
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

func (a *AuthenticatorOAuth2Introspection) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *AuthenticatorOAuth2Introspection) Config(config json.RawMessage) (*AuthenticatorOAuth2IntrospectionConfiguration, error) {
	var c AuthenticatorOAuth2IntrospectionConfiguration
	if err := a.c.AuthenticatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthenticatorMisconfigured(a, err)
	}

	if c.PreAuth != nil && c.PreAuth.Enabled {
		a.client = httpx.NewResilientClientLatencyToleranceSmall(
			(&clientcredentials.Config{
				ClientID:     c.PreAuth.ClientID,
				ClientSecret: c.PreAuth.ClientSecret,
				Scopes:       c.PreAuth.Scope,
				TokenURL:     c.PreAuth.TokenURL,
			}).
				Client(context.Background()).
				Transport,
		)
	}

	return &c, nil
}
