package authn

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/pipeline"
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
	Validate(config json.RawMessage) error
}

func NewErrAuthenticatorNotEnabled(a Authenticator) *herodot.DefaultError {
	return ErrAuthenticatorNotEnabled.WithTrace(errors.New("")).WithReasonf(`Authenticator "%s" is disabled per configuration.`, a.GetID())
}

func NewErrAuthenticatorMisconfigured(a Authenticator, err error) *herodot.DefaultError {
	return ErrAuthenticatorNotEnabled.WithTrace(err).WithReasonf(
		`Configuration for authenticator "%s" could not be validated: %s`,
		a.GetID(),
		err,
	)
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
