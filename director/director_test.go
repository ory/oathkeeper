package director

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"regexp"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/ory/oathkeeper/evaluator"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustCompileRegex(t *testing.T, pattern string) *regexp.Regexp {
	exp, err := regexp.Compile(pattern)
	require.NoError(t, err)
	return exp
}

func TestProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, helper.BearerTokenFromRequest(r))
		fmt.Fprint(w, r.Header.Get("Authorization"))
	}))
	defer backend.Close()

	u, _ := url.Parse(backend.URL)
	d := NewDirector(u, nil, nil, "some-secret")

	proxy := httptest.NewServer(&httputil.ReverseProxy{Director: d.Director, Transport: d})
	defer proxy.Close()

	publicRule := rule.Rule{MatchesMethods: []string{"GET"}, MatchesPathCompiled: mustCompileRegex(t, "/users/[0-9]+"), AllowAnonymous: true}
	disabledRule := rule.Rule{MatchesMethods: []string{"GET"}, MatchesPathCompiled: mustCompileRegex(t, "/users/[0-9]+"), BypassAuthorization: true}
	privateRule := rule.Rule{
		MatchesMethods:      []string{"GET"},
		MatchesPathCompiled: mustCompileRegex(t, "/users/([0-9]+)"),
		RequiredResource:    "users:$1",
		RequiredAction:      "get:$1",
		RequiredScopes:      []string{"users.create"},
	}

	for k, tc := range []struct {
		url       string
		code      int
		message   string
		rules     []rule.Rule
		mock      func(c *gomock.Controller) hydra.SDK
		transform func(r *http.Request)
		d         string
	}{
		{
			d:     "should fail because url does not exist in rule set",
			url:   proxy.URL + "/invalid",
			rules: []rule.Rule{},
			code:  http.StatusNotFound,
			mock: func(c *gomock.Controller) hydra.SDK {
				return nil
			},
		},
		{
			d:     "should fail because url does exist but is matched by two rules",
			url:   proxy.URL + "/users/1234",
			rules: []rule.Rule{publicRule, publicRule},
			code:  http.StatusInternalServerError,
			mock: func(c *gomock.Controller) hydra.SDK {
				return nil
			},
		},
		{
			d:     "should pass with an anonymous user because introspection fails but it doesn't matter because the endpoint is publicly available",
			url:   proxy.URL + "/users/1234",
			rules: []rule.Rule{publicRule},
			code:  http.StatusOK,
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := evaluator.NewMockSDK(c)
				s.EXPECT().IntrospectOAuth2Token(gomock.Eq("token"), gomock.Eq("")).Return(nil, nil, errors.New("error"))
				return s
			},
		},
		{
			d:     "should pass with an authorized user and a private rule",
			url:   proxy.URL + "/users/1234",
			rules: []rule.Rule{privateRule},
			code:  http.StatusOK,
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := evaluator.NewMockSDK(c)
				s.EXPECT().DoesWardenAllowTokenAccessRequest(gomock.Eq(swagger.WardenTokenAccessRequest{
					Token:    "token",
					Resource: "users:1234",
					Action:   "get:1234",
					Scopes:   []string{"users.create"},
					Context:  map[string]interface{}{"remoteIpAddress": "127.0.0.1"},
				})).Return(&swagger.WardenTokenAccessRequestResponse{Allowed: true}, &swagger.APIResponse{Response: &http.Response{StatusCode: http.StatusOK}}, nil)
				return s
			},
		},
		{
			d:       "should pass with a rule that bypasses authorization",
			url:     proxy.URL + "/users/1234",
			rules:   []rule.Rule{disabledRule},
			code:    http.StatusOK,
			message: "bearer token",
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := evaluator.NewMockSDK(c)
				return s
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			matcher := &rule.CachedMatcher{Rules: tc.rules}
			sdk := tc.mock(ctrl)
			d.Evaluator = evaluator.NewWardenEvaluator(nil, matcher, sdk)

			req, err := http.NewRequest("GET", tc.url, nil)
			require.NoError(t, err)
			if tc.transform != nil {
				tc.transform(req)
			}

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			greeting, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tc.code, res.StatusCode)
			if tc.message != "" {
				assert.Equal(t, tc.message, fmt.Sprintf("%s", greeting))
			}
		})
	}
}
