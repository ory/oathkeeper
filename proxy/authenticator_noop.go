package proxy

import (
	"net/http"
	"github.com/pkg/errors"
)

type AuthenticatorNoOp struct {}

func (a *AuthenticatorNoOp) Authenticate(r *http.Request) (*AuthenticationSession, error) {
	return nil, errors.WithStack(ErrAuthenticatorBypassed)
}
