package director

import (
	"net/url"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/hydra/sdk/go/hydra"
	"net/http"
	"github.com/pkg/errors"
	"context"
	"github.com/sirupsen/logrus"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/tomasen/realip"
	"github.com/ory/oathkeeper/helper"
)

func NewDirector(target *url.URL, sdk *hydra.SDK, matcher rule.Matcher, logger logrus.FieldLogger) *Director {
	if logger == nil {
		logger = logrus.New()
	}
	return &Director{
		TargetURL: target,
		SDK:       sdk,
		Matcher:   matcher,
		Logger:    logger,
	}
}

type Director struct {
	Matcher   rule.Matcher
	SDK       *hydra.SDK
	TargetURL *url.URL
	Logger    logrus.FieldLogger
}

type key int

const requestAllowed key = 0
const requestDenied key = 1

func (d *Director) RoundTrip(r *http.Request) (*http.Response, error) {
	if err, ok := r.Context().Value(requestDenied).(error); ok && err != nil {
		return &http.Response{
			StatusCode: http.StatusForbidden,
			//Body:       ioutil.NopCloser(bytes.NewBufferString(he.Description)),
		}, nil
	} else if token, ok := r.Context().Value(requestAllowed).(string); ok {
		r.Header.Set("Authorization", "bearer "+token)
		res, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			d.Logger.WithField("url", r.URL.String()).WithError(err).Print("Round trip failed.")
		}

		return res, err
	}

	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		//Body:       ioutil.NopCloser(bytes.NewBufferString(he.Description)),
	}, nil
}

func (d *Director) Director(r *http.Request) {

	access, err := d.GenerateWardenRequests(r)
	if errors.Cause(err) == helper.ErrPublicRule {
		token := helper.BearerTokenFromRequest(r)
		if token == "" {
			// public rule without authorization means access is allowed (anonymously)
			d.allowAnonymousRequest(r)
			return
		}

		result, resp, err := d.SDK.IntrospectOAuth2Token(token, "")
		if err != nil {
			// public rule with failing introspection is still valid
			d.Logger.WithField("url", r.URL.String()).WithError(err).Info("An error occurred during token introspection.")
			d.allowAnonymousRequest(r)
			return
		} else if resp.StatusCode != http.StatusOK {
			// public rule with failing introspection is still valid
			d.Logger.
				WithField("url", r.URL.String()).
				WithField("status_code", resp.StatusCode).
				Info("Token introspection did not result in HTTP status code 200.")
			d.allowAnonymousRequest(r)
			return
		} else if !result.Active {
			d.Logger.
				WithField("url", r.URL.String()).
				WithField("status_code", resp.StatusCode).
				Info("Token introspection says authorization bearer token inactive.")
			d.allowAnonymousRequest(r)
			return
		}

		d.allowIntrospectedRequest(r, result)
		return
	} else if err != nil {
		d.denyAnonymousRequest(r, err)
		return
	}

	result, resp, err := d.SDK.DoesWardenAllowTokenAccessRequest(access)
	if err != nil {
		d.denyAnonymousRequest(r, err)
	} else if resp.StatusCode != http.StatusOK {
		d.denyAnonymousRequest(r, err)
		return
	} else if !result.Allowed {
		d.denyAuthorizedRequest(r, result)
		return
	}
	d.allowAuthorizedRequest(r, result)
}

func (d *Director) GenerateAccessRequests(r *http.Request) ([]AccessRequest, error) {
	rules, err := d.Matcher.MatchRules(r.Method, r.URL)
	if err != nil {
		return nil, errors.WithStack(err)
	}

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
		if token == "" {
			return nil, errors.WithStack(helper.ErrMissingBearerToken)
		}

		access.Token = token
		access.Context = map[string]interface{}{
			"remoteIpAddress": realip.RealIP(r),
		}
		requests[k] = access
	}

	return requests, nil
}

func (d *Director) allowAnonymousRequest(r *http.Request) {
	d.Logger.WithFields(map[string]interface{}{"user": "anonymous", "url": r.URL.String()}).Info("Request granted.")
	r.URL.Scheme = d.TargetURL.Scheme
	r.URL.Host = d.TargetURL.Host
	*r = *r.WithContext(context.WithValue(r.Context(), requestAllowed, ""))
}

func (d *Director) allowIntrospectedRequest(r *http.Request, introspection *swagger.OAuth2TokenIntrospection) {
	d.Logger.WithFields(map[string]interface{}{"user": "anonymous", "url": r.URL.String()}).Info("Request granted.")
	r.URL.Scheme = d.TargetURL.Scheme
	r.URL.Host = d.TargetURL.Host
	*r = *r.WithContext(context.WithValue(r.Context(), requestAllowed, ""))
}

func (d *Director) denyAnonymousRequest(r *http.Request, err error) {
	d.Logger.WithError(err).WithFields(map[string]interface{}{"user": "anonymous", "url": r.URL.String()}).Info("Request denied.")
	*r = *r.WithContext(context.WithValue(r.Context(), requestDenied, err))
}

func (d *Director) denyAuthorizedRequest(r *http.Request, resp *swagger.WardenTokenAccessRequestResponsePayload) {
	*r = *r.WithContext(context.WithValue(r.Context(), requestDenied, ""))
}

func (d *Director) allowAuthorizedRequest(r *http.Request, resp *swagger.WardenTokenAccessRequestResponsePayload) {
	*r = *r.WithContext(context.WithValue(r.Context(), requestDenied, ""))
}
