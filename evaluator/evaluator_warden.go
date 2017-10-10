package evaluator

import (
	"net/http"

	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tomasen/realip"
)

type WardenEvaluator struct {
	Logger  logrus.FieldLogger
	Matcher rule.Matcher
	Hydra   hydra.SDK
}

func NewWardenEvaluator(l logrus.FieldLogger, m rule.Matcher, s hydra.SDK) *WardenEvaluator {
	if l == nil {
		l = logrus.New()
	}

	return &WardenEvaluator{
		Matcher: m,
		Hydra:   s,
		Logger:  l,
	}
}

func (d *WardenEvaluator) EvaluateAccessRequest(r *http.Request) (*Session, error) {
	token := helper.BearerTokenFromRequest(r)

	rl, err := d.Matcher.MatchRule(r.Method, r.URL)
	if err != nil {
		return nil, err
	}

	if rl.BypassAuthorization {
		return &Session{User: "", Anonymous: true, ClientID: "", Disabled: true}, nil
	}

	if rl.AllowAnonymous {
		if token == "" {
			return &Session{User: "", Anonymous: true, ClientID: ""}, nil
		}

		introspection, response, err := d.Hydra.IntrospectOAuth2Token(token, "")
		if err != nil {
			d.Logger.WithError(err).
				WithField("access_url", r.URL.String()).
				WithField("token", token[:5]).
				Errorf("Unable to connect to introspect endpoint.")
			return &Session{User: "", Anonymous: true, ClientID: ""}, nil
		} else if response.StatusCode != http.StatusOK {
			d.Logger.
				WithField("status_code", response.StatusCode).
				WithField("token", token[:5]).
				WithField("access_url", r.URL.String()).
				Errorf("Expected introspection response to return status code 200.")
			return &Session{User: "", Anonymous: true, ClientID: ""}, nil
		} else if !introspection.Active {
			return &Session{User: "", Anonymous: true, ClientID: ""}, nil
		}

		return &Session{
			User:      introspection.Sub,
			ClientID:  introspection.ClientId,
			Anonymous: false,
		}, nil
	}

	if token == "" {
		return nil, errors.WithStack(helper.ErrMissingBearerToken)
	}

	introspection, response, err := d.Hydra.DoesWardenAllowTokenAccessRequest(d.prepareAccessRequests(r, token, rl))
	if err != nil {
		d.Logger.WithError(err).
			WithField("access_url", r.URL.String()).
			WithField("token", token[:5]).
			Errorf("Unable to connect to warden endpoint.")
		return nil, errors.WithStack(err)
	} else if response.StatusCode != http.StatusOK {
		d.Logger.
			WithField("status_code", response.StatusCode).
			WithField("token", token[:5]).
			WithField("access_url", r.URL.String()).
			Errorf("Expected warden response to return status code 200.")
		return nil, errors.Errorf("Token introspection expects status code %d but got %d", http.StatusOK, response.StatusCode)
	} else if !introspection.Allowed {
		return nil, errors.WithStack(helper.ErrForbidden)
	}

	return &Session{
		User:      introspection.Subject,
		ClientID:  introspection.ClientId,
		Anonymous: false,
	}, nil
}

func (d *WardenEvaluator) prepareAccessRequests(r *http.Request, token string, rl *rule.Rule) swagger.WardenTokenAccessRequest {
	return swagger.WardenTokenAccessRequest{
		Scopes:   rl.RequiredScopes,
		Action:   rl.MatchesPath.ReplaceAllString(r.URL.Path, rl.RequiredAction),
		Resource: rl.MatchesPath.ReplaceAllString(r.URL.Path, rl.RequiredResource),
		Token:    token,
		Context: map[string]interface{}{
			"remoteIpAddress": realip.RealIP(r),
		},
	}
}
