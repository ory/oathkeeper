package decision

import (
	"github.com/ory/oathkeeper/rule"
	"net/http"
)

type Decider interface {
	AllowAccess(r *http.Request, rules []rule.Rule) (*Session, error)
}

type Session struct {
}
