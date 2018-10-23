package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ory/fosite"
	"github.com/ory/go-convenience/stringslice"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/clientcredentials"
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
	client           *http.Client
	introspectionURL string
	scopeStrategy    fosite.ScopeStrategy
}

func NewAuthenticatorOAuth2Introspection(
	clientID, clientSecret, tokenURL, introspectionURL string,
	scopes []string, strategy fosite.ScopeStrategy,
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

func (a *AuthenticatorOAuth2Introspection) Authenticate(r *http.Request, config json.RawMessage, rl *rule.Rule) (*AuthenticationSession, error) {
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

	var ir swagger.OAuth2TokenIntrospection
	if err := json.NewDecoder(resp.Body).Decode(&ir); err != nil {
		return nil, errors.WithStack(err)
	}

	if len(ir.TokenType) > 0 && ir.TokenType != "access_token" {
		return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Introspected token is not an access token but \"%s\"", ir.TokenType)))
	}

	if !ir.Active {
		return nil, errors.WithStack(helper.ErrForbidden.WithReason("Access token introspection says token is not active"))
	}

	for _, audience := range cf.Audience {
		if !stringslice.Has(ir.Aud, audience) {
			return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Token audience is not intended for target audience %s", audience)))
		}
	}

	if len(cf.Issuers) > 0 {
		if !stringslice.Has(cf.Issuers, ir.Iss) {
			return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Token issuer does not match any trusted issuer")))
		}
	}

	if a.scopeStrategy != nil {
		for _, scope := range cf.Scopes {
			if !a.scopeStrategy(strings.Split(ir.Scope, " "), scope) {
				return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Scope %s was not granted", scope)))
			}
		}
	}

	if len(ir.Ext) == 0 {
		ir.Ext = map[string]interface{}{}
	}

	ir.Ext["username"] = ir.Username
	ir.Ext["client_id"] = ir.ClientId
	ir.Ext["scope"] = ir.Scope

	return &AuthenticationSession{
		Subject: ir.Sub,
		Extra:   ir.Ext,
	}, nil
}
