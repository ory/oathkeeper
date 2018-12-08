package proxy

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/rule"
)

type AuthenticatorAnonymous struct {
	AnonymousIdentifier string
}

func NewAuthenticatorAnonymous(anonymousIdentifier string) *AuthenticatorAnonymous {
	return &AuthenticatorAnonymous{
		AnonymousIdentifier: anonymousIdentifier,
	}
}

func (a *AuthenticatorAnonymous) GetID() string {
	return "anonymous"
}

func (a *AuthenticatorAnonymous) Authenticate(r *http.Request, config json.RawMessage, rl *rule.Rule) (*AuthenticationSession, error) {
	if len(r.Header.Get("Authorization")) != 0 {
		return nil, errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	return &AuthenticationSession{Subject: a.AnonymousIdentifier}, nil
}
