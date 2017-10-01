package decision

import (
	"github.com/ory/oathkeeper/rule"
	"net/http"
	"github.com/tomasen/realip"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/ory/oathkeeper/helper"
	"github.com/pkg/errors"
)

type WardenDecider struct{}

const (
	statusForceDeny = 2
	statusDefaultDeny = 0
	statusAllow = 1
)

func (d *WardenDecider) AllowAccess(r *http.Request, rules []rule.Rule) (*Session, error) {
	token := helper.BearerTokenFromRequest(r)
f


	if token == "" {
		return nil, errors.WithStack(helper.ErrMissingBearerToken)
	}



	requests, err := d.PrepareAccessRequests(r,rules)
	status := statusDefaultDeny
	var queryRequests []AccessRequest

	for _, access := range requests {
		if access.Public {
			status = statusAllow
			continue
		}
	}
}

func (d *WardenDecider) PrepareAccessRequests(r *http.Request, rules []rule.Rule) ([]AccessRequest, error) {
	requests := make([]AccessRequest, len(rules))
	for k, matched := range rules {
		access := AccessRequest{
			WardenTokenAccessRequest: swagger.WardenTokenAccessRequest{
				Scopes:   matched.RequiredScopes,
				Action:   matched.MatchesPath.ReplaceAllString(r.URL.Path, matched.RequiredAction),
				Resource: matched.MatchesPath.ReplaceAllString(r.URL.Path, matched.RequiredResource),
			},
			Public: matched.Public,
		}

		token := helper.BearerTokenFromRequest(r)
		access.Token = token
		access.Context = map[string]interface{}{
			"remoteIpAddress": realip.RealIP(r),
		}
		requests[k] = access
	}

	return requests, nil
}