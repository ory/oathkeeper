package proxy

import (
	"net/http"
	"github.com/go-errors/errors"
)

var ErrAuthenticatorNotResponsible = errors.New("Authenticator not responsible")
var ErrAuthenticatorBypassed = errors.New("Authenticator is disabled")

type Authenticator interface {
	Authenticate(r *http.Request) (*AuthenticationSession, error)
}

type AuthenticationSession struct {
}
