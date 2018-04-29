package proxy

import (
	"encoding/json"
	"github.com/ory/oathkeeper/rule"
	"net/http"
)

type Authorizer interface {
	Authorize(r *http.Request, session *AuthenticationSession, config json.RawMessage, rl *rule.Rule) error
	GetID() string
}

// This field will be used to decide advanced authorization requests where access control policies are used. A
// action is typically something a user wants to do (e.g. write, read, delete).
// This field supports expansion as described in the developer guide: https://ory.gitbooks.io/oathkeeper/content/concepts.html#rules
//RequiredAction string `json:"requiredAction"`

// This field will be used to decide advanced authorization requests where access control policies are used. A
// resource is typically something a user wants to access (e.g. printer, article, virtual machine).
// This field supports expansion as described in the developer guide: https://ory.gitbooks.io/oathkeeper/content/concepts.html#rules
//RequiredResource string `json:"requiredResource"`
