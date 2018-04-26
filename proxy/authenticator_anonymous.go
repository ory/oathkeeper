package proxy

import (
	"net/http"
	"github.com/pkg/errors"
)

type AuthenticatorAnonymous struct {
	AnonymousIdentifier string
}

func NewAuthenticatorAnonymous(anonymousIdentifier string) *AuthenticatorAnonymous {
	return &AuthenticatorAnonymous{
		AnonymousIdentifier: anonymousIdentifier,
	}
}

func (a *AuthenticatorAnonymous) Authenticate(r *http.Request) (*AuthenticationSession, error) {
	if len(r.Header.Get("Authorization")) != 0 {
		return nil, errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	return &AuthenticationSession{}, nil
}
