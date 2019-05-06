package authz

import (
	"encoding/json"
	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/pipeline/authn"
	"net/http"

	"github.com/ory/oathkeeper/rule"
)

var ErrAuthorizerNotEnabled = herodot.DefaultError{
	ErrorField: "authorizer matching this route is misconfigured or disabled",
	CodeField:   http.StatusInternalServerError,
	StatusField: http.StatusText(http.StatusInternalServerError),
}

type Authorizer interface {
	Authorize(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rl *rule.Rule) error
	GetID() string
	Validate() error
}
