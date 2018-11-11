package proxy

import (
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/rule"
)

type Authorizer interface {
	Authorize(r *http.Request, session *AuthenticationSession, config json.RawMessage, rl *rule.Rule) error
	GetID() string
}
