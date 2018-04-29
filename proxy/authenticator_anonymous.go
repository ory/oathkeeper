package proxy

import (
	"encoding/json"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"net/http"
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
