package proxy

import (
	"encoding/json"
	"github.com/ory/herodot"
	"net/http"

	"github.com/go-errors/errors"

	"github.com/ory/oathkeeper/rule"
)

var ErrAuthenticatorNotResponsible = errors.New("Authenticator not responsible")
var ErrAuthenticatorNotEnabled = herodot.DefaultError{
	ErrorField: "authenticator matching this route is misconfigured or disabled",
	CodeField:   http.StatusInternalServerError,
	StatusField: http.StatusText(http.StatusInternalServerError),
}

type Authenticator interface {
	Authenticate(r *http.Request, config json.RawMessage, rl *rule.Rule) (*AuthenticationSession, error)
	GetID() string
	Validate() error
}

type AuthenticationSession struct {
	Subject string
	Extra   map[string]interface{}
}
