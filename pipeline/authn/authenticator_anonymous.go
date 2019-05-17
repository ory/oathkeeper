package authn

import (
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"

	"github.com/pkg/errors"
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
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf(`Authenticator "%s" is disabled per configuration.`, a.GetID()))
	}

	return nil
}

func (a *AuthenticatorAnonymous) GetID() string {
	return "anonymous"
}

func (a *AuthenticatorAnonymous) Authenticate(r *http.Request, config json.RawMessage, _ pipeline.Rule) (*AuthenticationSession, error) {
	if len(r.Header.Get("Authorization")) != 0 {
		return nil, errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	return &AuthenticationSession{Subject: a.c.AuthenticatorAnonymousIdentifier()}, nil
}
