package extauth

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/ory/oathkeeper/evaluator"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func mustCompileRegex(t *testing.T, pattern string) *regexp.Regexp {
	exp, err := regexp.Compile(pattern)
	require.NoError(t, err)
	return exp
}

func TestExtAuth(t *testing.T) {
	url := "http://original.localhost.local"

	handler := Handler{Evaluator: nil}
	router := httprouter.New()
	handler.SetRoutes(router)
	server := httptest.NewServer(router)

	publicRule := rule.Rule{MatchesMethods: []string{"GET"}, MatchesURLCompiled: mustCompileRegex(t, url+"/users/[0-9]+"), Mode: rule.AnonymousMode}
	disabledRule := rule.Rule{MatchesMethods: []string{"GET"}, MatchesURLCompiled: mustCompileRegex(t, url+"/users/[0-9]+"), Mode: rule.BypassMode}
	privateRule := rule.Rule{
		MatchesMethods:     []string{"GET"},
		MatchesURLCompiled: mustCompileRegex(t, url+"/users/([0-9]+)"),
		RequiredResource:   "users:$1",
		RequiredAction:     "get:$1",
		RequiredScopes:     []string{"users.create"},
		Mode:               rule.PolicyMode,
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
			d:     "should fail because x-original-url missing in header",
			rules: []rule.Rule{},
			code:  http.StatusBadRequest,
			transform: func(r *http.Request) {
				r.Header.Add("x-original-method", "GET")
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				return nil
			},
		},
		{
			d:     "should fail because x-original-method missing in header",
			rules: []rule.Rule{},
			code:  http.StatusBadRequest,
			transform: func(r *http.Request) {
				r.Header.Add("x-original-url", url+"/users/1234")
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				return nil
			},
		},
		{
			d:     "should fail because url does not exist in rule set",
			rules: []rule.Rule{},
			code:  http.StatusNotFound,
			transform: func(r *http.Request) {
				r.Header.Add("x-original-url", url+"/invalid")
				r.Header.Add("x-original-method", "GET")
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				return nil
			},
		},
		{
			d:     "should fail because url does exist but is matched by two rules",
			rules: []rule.Rule{publicRule, publicRule},
			code:  http.StatusInternalServerError,
			transform: func(r *http.Request) {
				r.Header.Add("x-original-url", url+"/users/1234")
				r.Header.Add("x-original-method", "GET")
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				return nil
			},
		},
		{
			d:     "should pass with an anonymous user because introspection fails but it doesn't matter because the endpoint is publicly available",
			rules: []rule.Rule{publicRule},
			code:  http.StatusOK,
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
				r.Header.Add("x-original-url", url+"/users/1234")
				r.Header.Add("x-original-method", "GET")
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := evaluator.NewMockSDK(c)
				s.EXPECT().IntrospectOAuth2Token(gomock.Eq("token"), gomock.Eq("")).Return(nil, nil, errors.New("error"))
				return s
			},
		},
		{
			d:     "should pass with an authorized user and a private rule",
			rules: []rule.Rule{privateRule},
			code:  http.StatusOK,
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
				r.Header.Add("x-original-url", url+"/users/1234")
				r.Header.Add("x-original-method", "GET")
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
			d:     "should pass with a rule that bypasses authorization",
			rules: []rule.Rule{disabledRule},
			code:  http.StatusOK,
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
				r.Header.Add("x-original-url", url+"/users/1234")
				r.Header.Add("x-original-method", "GET")
			},
			mock: func(c *gomock.Controller) hydra.SDK {
				s := evaluator.NewMockSDK(c)
				return s
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			matcher := &rule.CachedMatcher{Rules: tc.rules}
			sdk := tc.mock(ctrl)

			handler.Evaluator = evaluator.NewWardenEvaluator(nil, matcher, sdk, "")

			req, err := http.NewRequest("GET", server.URL+"/extauth", nil)
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

func panicCompileRegex(pattern string) *regexp.Regexp {
	exp, err := regexp.Compile(pattern)
	if err != nil {
		panic(err.Error())
	}
	return exp
}

func BenchmarkExtauth(b *testing.B) {
	logger := logrus.New()
	logger.Level = logrus.WarnLevel

	url := "http://original.localhost.local"

	handler := Handler{}
	router := httprouter.New()
	handler.SetRoutes(router)
	server := httptest.NewServer(router)

	matcher := &rule.CachedMatcher{Rules: []rule.Rule{
		{MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(url + "/users"), Mode: rule.AnonymousMode},
		{MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(url + "/users/<[0-9]+>"), Mode: rule.AnonymousMode},
		{MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(url + "/<[0-9]+>"), Mode: rule.AnonymousMode},
		{MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(url + "/other/<.+>"), Mode: rule.AnonymousMode},
	}}
	handler.Evaluator = evaluator.NewWardenEvaluator(logger, matcher, nil, "")

	req, _ := http.NewRequest("GET", server.URL+"/extauth", nil)
	req.Header.Add("x-original-url", url+"/users")
	req.Header.Add("x-original-method", "GET")

	b.Run("case=fetch_user_endpoint", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				b.FailNow()
			}

			if res.StatusCode != http.StatusOK {
				b.FailNow()
			}
		}
	})
}
