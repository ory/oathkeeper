package proxy

import (
	"net/http"
	"github.com/pkg/errors"
	"encoding/json"
	"github.com/ory/keto/authentication"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/fosite"
	"bytes"
	"fmt"
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
	helper           authenticatorOAuth2IntrospectionHelper
	introspectionURL string
	scopeStrategy    fosite.ScopeStrategy
}

type authenticatorOAuth2IntrospectionHelper interface {
	Introspect(token string, scopes []string, strategy fosite.ScopeStrategy) (*authentication.IntrospectionResponse, error)
}

func NewAuthenticatorOAuth2Introspection(clientID, clientSecret, tokenURL, introspectionURL string, scopes []string, strategy fosite.ScopeStrategy) *AuthenticatorOAuth2Introspection {
	return &AuthenticatorOAuth2Introspection{
		helper:           authentication.NewOAuth2IntrospectionAuthentication(clientID, clientSecret, tokenURL, introspectionURL, scopes, strategy),
		introspectionURL: introspectionURL,
	}
}

func (a *AuthenticatorOAuth2Introspection) GetID() string {
	return "oauth2_introspection"
}

func (a *AuthenticatorOAuth2Introspection) Authenticate(r *http.Request, config json.RawMessage) (*AuthenticationSession, error) {
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

	ir, err := a.helper.Introspect(token, cf.Scopes, a.scopeStrategy)
	if err != nil {
		return nil, err
	}

	for _, audience := range cf.Audience {
		if !stringInSlice(audience, ir.Audience) {
			return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Token audience is not intended for target audience %s", audience)))
		}
	}

	if len(cf.Issuers) > 0 {
		if !stringInSlice(ir.Issuer, cf.Issuers) {
			return nil, errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Token issuer does not match any trusted issuer")))
		}
	}

	ir.Extra["username"] = ir.Username
	ir.Extra["client_id"] = ir.ClientID
	ir.Extra["scope"] = ir.Scope

	return &AuthenticationSession{
		Subject: ir.Subject,
		Extra:   ir.Extra,
	}, nil
}

func stringInSlice(needle string, haystack []string) bool {
	for _, b := range haystack {
		if b == needle {
			return true
		}
	}
	return false
}
