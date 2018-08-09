package proxy

import (
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/rule"
)

type AuthenticatorNoOp struct{}

func NewAuthenticatorNoOp() *AuthenticatorNoOp {
	return new(AuthenticatorNoOp)
}

func (a *AuthenticatorNoOp) GetID() string {
	return "noop"
}

func (a *AuthenticatorNoOp) Authenticate(r *http.Request, config json.RawMessage, rl *rule.Rule) (*AuthenticationSession, error) {
	return &AuthenticationSession{Subject: ""}, nil
}
