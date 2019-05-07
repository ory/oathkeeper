package authz

import (
	"encoding/json"
	"net/http"

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
)

var ErrAuthorizerNotEnabled = herodot.DefaultError{
	ErrorField:  "authorizer matching this route is misconfigured or disabled",
	CodeField:   http.StatusInternalServerError,
	StatusField: http.StatusText(http.StatusInternalServerError),
}

type Authorizer interface {
	Authorize(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error
	GetID() string
	Validate() error
}
