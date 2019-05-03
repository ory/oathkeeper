package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ory/oathkeeper/driver/configuration"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/fosite"
	"github.com/ory/go-convenience/stringslice"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
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

	client           *http.Client
	introspectionURL string
	scopeStrategy    fosite.ScopeStrategy
}

func NewAuthenticatorOAuth2Introspection(
	c configuration.Provider,
) (*AuthenticatorOAuth2Introspection, error) {
	if _, err := url.ParseRequestURI(introspectionURL); err != nil {
		return new(AuthenticatorOAuth2Introspection), errors.Errorf(`unable to validate the OAuth 2.0 Introspection Authenticator's Token Introspection URL "%s" because %s`, introspectionURL, err)
	}

	c := http.DefaultClient
	if len(clientID)+len(clientSecret)+len(tokenURL)+len(scopes) > 0 {
		if len(clientID) == 0 {
			return new(AuthenticatorOAuth2Introspection), errors.Errorf("if OAuth 2.0 Authorization is used in the OAuth 2.0 Introspection Authenticator, the OAuth 2.0 Client ID must be set but was not")
		}
		if len(clientSecret) == 0 {
			return new(AuthenticatorOAuth2Introspection), errors.Errorf("if OAuth 2.0 Authorization is used in the OAuth 2.0 Introspection Authenticator, the OAuth 2.0 Client ID must be set but was not")
		}
		if _, err := url.ParseRequestURI(tokenURL); err != nil {
			return new(AuthenticatorOAuth2Introspection), errors.Errorf(`if OAuth 2.0 Authorization is used in the OAuth 2.0 Introspection Authenticator, the OAuth 2.0 Token URL must be set but validating URL "%s" failed because %s`, tokenURL, err)
		}

		c = (&clientcredentials.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			TokenURL:     tokenURL,
			Scopes:       scopes,
		}).Client(context.Background())
	}

	return &AuthenticatorOAuth2Introspection{
		client:           c,
		introspectionURL: introspectionURL,
		scopeStrategy:    strategy,
	}, nil
}

func (a *AuthenticatorOAuth2Introspection) GetID() string {
	return "oauth2_introspection"
}

type introspection struct {
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

func (a *AuthenticatorOAuth2Introspection) Authenticate(r *http.Request, config json.RawMessage, rl *rule.Rule) (*AuthenticationSession, error) {
	var i introspection
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
	resp, err := a.client.Post(a.introspectionURL, "application/x-www-form-urlencoded", strings.NewReader(body.Encode()))
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

	if a.scopeStrategy != nil {
		for _, scope := range cf.Scopes {
			if !a.scopeStrategy(strings.Split(i.Scope, " "), scope) {
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
