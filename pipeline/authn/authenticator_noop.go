package authn

import (
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
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

func (a *AuthenticatorNoOp) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	if err := a.c.AuthenticatorConfig(a.GetID(), config, nil); err != nil {
		return NewErrAuthenticatorMisconfigured(a, err)
	}
	return nil
}

func (a *AuthenticatorNoOp) Authenticate(r *http.Request, config json.RawMessage, _ pipeline.Rule) (*AuthenticationSession, error) {
	return &AuthenticationSession{Subject: ""}, nil
}
