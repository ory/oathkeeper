package proxy

import (
	"encoding/json"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/pkg/errors"
	"net/http"

	"github.com/ory/oathkeeper/rule"
)

type AuthenticatorNoOp struct {
	c configuration.Provider
}

func NewAuthenticatorNoOp(c configuration.Provider) *AuthenticatorNoOp {
	return &AuthenticatorNoOp{c: c}
}

func (a *AuthenticatorNoOp) GetID() string {
	return "noop"
}

func (a *AuthenticatorNoOp) Validate() error {
	if !a.c.AuthenticatorNoopIsEnabled() {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf("Authenticator % is disabled per configuration.", a.GetID()))
	}

	return nil
}

func (a *AuthenticatorNoOp) Authenticate(r *http.Request, config json.RawMessage, rl *rule.Rule) (*AuthenticationSession, error) {
	return &AuthenticationSession{Subject: ""}, nil
}
