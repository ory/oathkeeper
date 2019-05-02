package proxy

import (
	"encoding/json"
	"net/http"

	"github.com/go-errors/errors"

	"github.com/ory/oathkeeper/rule"
)

var ErrAuthenticatorNotResponsible = errors.New("Authenticator not responsible")
var ErrAuthenticatorBypassed = errors.New("Authenticator is disabled")

type Authenticator interface {
	Authenticate(r *http.Request, config json.RawMessage, rl *rule.Rule) (*AuthenticationSession, error)
	GetID() string
}

type AuthenticationSession struct {
	Subject string
	Extra   map[string]interface{}
}
