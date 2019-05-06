package authn

import (
	"encoding/json"
	"github.com/ory/oathkeeper/driver/configuration"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/rule"
)

type AuthenticatorAnonymous struct {
	c configuration.Provider
}

func NewAuthenticatorAnonymous(c configuration.Provider) *AuthenticatorAnonymous {
	return &AuthenticatorAnonymous{
		c: c,
	}
}

func (a *AuthenticatorAnonymous) Validate() error {
	if !a.c.AuthenticatorAnonymousIsEnabled() {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf("Authenticator % is disabled per configuration.", a.GetID()))
	}

	return nil
}

func (a *AuthenticatorAnonymous) GetID() string {
	return "anonymous"
}

func (a *AuthenticatorAnonymous) Authenticate(r *http.Request, config json.RawMessage, rl *rule.Rule) (*AuthenticationSession, error) {
	if len(r.Header.Get("Authorization")) != 0 {
		return nil, errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	return &AuthenticationSession{Subject: a.c.AuthenticatorAnonymousIdentifier()}, nil
}
