package authn

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

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

func (a *AuthenticatorNoOp) Validate() error {
	if !a.c.AuthenticatorNoopIsEnabled() {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf(`Authenticator "%s" is disabled per configuration.`, a.GetID()))
	}

	return nil
}

func (a *AuthenticatorNoOp) Authenticate(r *http.Request, config json.RawMessage, _ pipeline.Rule) (*AuthenticationSession, error) {
	return &AuthenticationSession{Subject: ""}, nil
}
