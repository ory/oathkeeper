package proxy

import (
	"net/http"
	"github.com/pkg/errors"
	"encoding/json"
)

type AuthenticatorNoOp struct{}

func NewAuthenticatorNoOp() *AuthenticatorNoOp {
	return new(AuthenticatorNoOp)
}

func (a *AuthenticatorNoOp) GetID() string {
	return "noop"
}

func (a *AuthenticatorNoOp) Authenticate(r *http.Request, config json.RawMessage) (*AuthenticationSession, error) {
	return nil, errors.WithStack(ErrAuthenticatorBypassed)
}
