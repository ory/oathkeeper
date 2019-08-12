package authn

import (
	"encoding/json"
	"net/http"

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/pipeline"

	"github.com/go-errors/errors"
)

var ErrAuthenticatorNotResponsible = errors.New("Authenticator not responsible")
var ErrAuthenticatorNotEnabled = herodot.DefaultError{
	ErrorField:  "authenticator matching this route is misconfigured or disabled",
	CodeField:   http.StatusInternalServerError,
	StatusField: http.StatusText(http.StatusInternalServerError),
}

type Authenticator interface {
	Authenticate(r *http.Request, config json.RawMessage, rule pipeline.Rule) (*AuthenticationSession, error)
	GetID() string
	Validate() error
}

type AuthenticationSession struct {
	Subject string                 `json:"subject"`
	Extra   map[string]interface{} `json:"extra"`
	Header  http.Header            `json:"header"`
}

func (a *AuthenticationSession) SetHeader(key, val string) {
	if a.Header == nil {
		a.Header = map[string][]string{}
	}
	a.Header.Set(key, val)
}
