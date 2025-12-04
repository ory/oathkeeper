// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"

	"github.com/ory/herodot"
	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/api"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rule"
)

func TestDecisionAPI(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	conf.SetForTest(t, configuration.AuthenticatorNoopIsEnabled, true)
	conf.SetForTest(t, configuration.AuthenticatorUnauthorizedIsEnabled, true)
	conf.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, true)
	conf.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
	conf.SetForTest(t, configuration.AuthorizerDenyIsEnabled, true)
	conf.SetForTest(t, configuration.MutatorNoopIsEnabled, true)
	conf.SetForTest(t, configuration.ErrorsWWWAuthenticateIsEnabled, true)
	reg := internal.NewRegistry(conf).WithBrokenPipelineMutator()

	d := reg.DecisionHandler()

	n := negroni.New(d)
	n.UseHandler(httprouter.New())

	ts := httptest.NewServer(n)
	defer ts.Close()

	ruleNoOpAuthenticator := rule.Rule{
		Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-noop/<[0-9]+>"},
		Authenticators: []rule.Handler{{Handler: "noop"}},
		Authorizer:     rule.Handler{Handler: "allow"},
		Mutators:       []rule.Handler{{Handler: "noop"}},
		Upstream:       rule.Upstream{URL: ""},
	}
	ruleNoOpAuthenticatorModifyUpstream := rule.Rule{
		Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/strip-path/authn-noop/<[0-9]+>"},
		Authenticators: []rule.Handler{{Handler: "noop"}},
		Authorizer:     rule.Handler{Handler: "allow"},
		Mutators:       []rule.Handler{{Handler: "noop"}},
		Upstream:       rule.Upstream{URL: "", StripPath: "/strip-path/", PreserveHost: true},
	}

	ruleNoOpAuthenticatorGLOB := rule.Rule{
		Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-noop/<[0-9]*>"},
		Authenticators: []rule.Handler{{Handler: "noop"}},
		Authorizer:     rule.Handler{Handler: "allow"},
		Mutators:       []rule.Handler{{Handler: "noop"}},
		Upstream:       rule.Upstream{URL: ""},
	}
	ruleNoOpAuthenticatorModifyUpstreamGLOB := rule.Rule{
		Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/strip-path/authn-noop/<[0-9]*>"},
		Authenticators: []rule.Handler{{Handler: "noop"}},
		Authorizer:     rule.Handler{Handler: "allow"},
		Mutators:       []rule.Handler{{Handler: "noop"}},
		Upstream:       rule.Upstream{URL: "", StripPath: "/strip-path/", PreserveHost: true},
	}

	for k, tc := range []struct {
		url         string
		code        int
		reqBody     []byte
		rulesRegexp []rule.Rule
		rulesGlob   []rule.Rule
		transform   func(r *http.Request)
		authz       string
		d           string
	}{
		{
			d:           "should fail because url does not exist in rule set",
			url:         ts.URL + "/decisions" + "/invalid",
			rulesRegexp: []rule.Rule{},
			rulesGlob:   []rule.Rule{},
			code:        http.StatusNotFound,
		},
		{
			d:           "should fail because url does exist but is matched by two rulesRegexp",
			url:         ts.URL + "/decisions" + "/authn-noop/1234",
			rulesRegexp: []rule.Rule{ruleNoOpAuthenticator, ruleNoOpAuthenticator},
			rulesGlob:   []rule.Rule{ruleNoOpAuthenticatorGLOB, ruleNoOpAuthenticatorGLOB},
			code:        http.StatusInternalServerError,
		},
		{
			d:           "should pass",
			url:         ts.URL + "/decisions" + "/authn-noop/1234",
			rulesRegexp: []rule.Rule{ruleNoOpAuthenticator},
			rulesGlob:   []rule.Rule{ruleNoOpAuthenticatorGLOB},
			code:        http.StatusOK,
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
			},
			authz: "bearer token",
		},
		{
			d:           "should pass",
			url:         ts.URL + "/decisions" + "/strip-path/authn-noop/1234",
			rulesRegexp: []rule.Rule{ruleNoOpAuthenticatorModifyUpstream},
			rulesGlob:   []rule.Rule{ruleNoOpAuthenticatorModifyUpstreamGLOB},
			code:        http.StatusOK,
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
			},
			authz: "bearer token",
		},
		{
			d:   "should fail because no authorizer was configured",
			url: ts.URL + "/decisions" + "/authn-anon/authz-none/cred-none/1234",
			rulesRegexp: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-none/cred-none/<[0-9]+>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			rulesGlob: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-none/cred-none/<[0-9]*>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
			},
			code: http.StatusUnauthorized,
		},
		{
			d:   "should fail because no mutator was configured",
			url: ts.URL + "/decisions" + "/authn-anon/authz-allow/cred-none/1234",
			rulesRegexp: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-allow/cred-none/<[0-9]+>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Upstream:       rule.Upstream{URL: ""},
			}},
			rulesGlob: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-allow/cred-none/<[0-9]*>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Upstream:       rule.Upstream{URL: ""},
			}},
			code: http.StatusInternalServerError,
		},
		{
			d:   "should pass with anonymous and everything else set to noop",
			url: ts.URL + "/decisions" + "/authn-anon/authz-allow/cred-noop/1234",
			rulesRegexp: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-allow/cred-noop/<[0-9]+>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			rulesGlob: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-allow/cred-noop/<[0-9]*>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			code:  http.StatusOK,
			authz: "",
		},
		{
			d:   "should fail when authorizer fails",
			url: ts.URL + "/decisions" + "/authn-anon/authz-deny/cred-noop/1234",
			rulesRegexp: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-deny/cred-noop/<[0-9]+>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "deny"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			rulesGlob: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-deny/cred-noop/<[0-9]*>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "deny"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			code: http.StatusForbidden,
		},
		{
			d:   "should fail when authenticator fails",
			url: ts.URL + "/decisions" + "/authn-broken/authz-none/cred-none/1234",
			rulesRegexp: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-broken/authz-none/cred-none/<[0-9]+>"},
				Authenticators: []rule.Handler{{Handler: "unauthorized"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			rulesGlob: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-broken/authz-none/cred-none/<[0-9]*>"},
				Authenticators: []rule.Handler{{Handler: "unauthorized"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			code: http.StatusUnauthorized,
		},
		{
			d:   "should fail when mutator fails",
			url: ts.URL + "/decisions" + "/authn-anonymous/authz-allow/cred-broken/1234",
			rulesRegexp: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anonymous/authz-allow/cred-broken/<[0-9]+>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "broken"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			rulesGlob: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anonymous/authz-allow/cred-broken/<[0-9]*>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "broken"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			code: http.StatusInternalServerError,
		},
		{
			d:   "should fail when one of the mutators fails",
			url: ts.URL + "/decisions" + "/authn-anonymous/authz-allow/cred-broken/1234",
			rulesRegexp: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anonymous/authz-allow/cred-broken/<[0-9]+>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}, {Handler: "broken"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			rulesGlob: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anonymous/authz-allow/cred-broken/<[0-9]*>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}, {Handler: "broken"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			code: http.StatusInternalServerError,
		},
		{
			d:   "should fail when authorizer fails and send www_authenticate as defined in the rule",
			url: ts.URL + "/decisions" + "/authn-anon/authz-deny/cred-noop/1234",
			rulesRegexp: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-deny/cred-noop/<[0-9]+>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "deny"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: ""},
				Errors:         []rule.ErrorHandler{{Handler: "www_authenticate"}},
			}},
			rulesGlob: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-deny/cred-noop/<[0-9]*>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "deny"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: ""},
				Errors:         []rule.ErrorHandler{{Handler: "www_authenticate"}},
			}},
			code: http.StatusUnauthorized,
		},
		{
			d:   "should not pass Content-Length from client",
			url: ts.URL + "/decisions" + "/authn-anon/authz-allow/cred-noop/1234",
			rulesRegexp: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-allow/cred-noop/<[0-9]+>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			rulesGlob: []rule.Rule{{
				Match:          &rule.Match{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-allow/cred-noop/<[0-9]*>"},
				Authenticators: []rule.Handler{{Handler: "anonymous"}},
				Authorizer:     rule.Handler{Handler: "allow"},
				Mutators:       []rule.Handler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			reqBody: []byte("non-empty body"),
			transform: func(r *http.Request) {
				r.Header.Add("Content-Length", "1337")
			},
			code:  http.StatusOK,
			authz: "",
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
			testFunc := func(strategy configuration.MatchingStrategy) {
				require.NoError(t, reg.RuleRepository().SetMatchingStrategy(context.Background(), strategy))
				req, err := http.NewRequest("GET", tc.url, bytes.NewBuffer(tc.reqBody))
				require.NoError(t, err)
				if tc.transform != nil {
					tc.transform(req)
				}

				res, err := http.DefaultClient.Do(req)
				require.NoError(t, err)

				entireBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				defer res.Body.Close() //nolint:errcheck // closing test response

				assert.Equal(t, tc.authz, res.Header.Get("Authorization"))
				assert.Equal(t, tc.code, res.StatusCode)
				assert.Equal(t, strconv.Itoa(len(entireBody)), res.Header.Get("Content-Length"))
			}
			t.Run("regexp", func(_ *testing.T) {
				reg.RuleRepository().(*rule.RepositoryMemory).WithRules(tc.rulesRegexp)
				testFunc(configuration.Regexp)
			})
			t.Run("glob", func(_ *testing.T) {
				reg.RuleRepository().(*rule.RepositoryMemory).WithRules(tc.rulesGlob)
				testFunc(configuration.Glob)
			})
		})
	}
}

type decisionHandlerRegistryMock struct {
	mock.Mock
}

func (m *decisionHandlerRegistryMock) RuleMatcher() rule.Matcher {
	return m
}

func (m *decisionHandlerRegistryMock) ProxyRequestHandler() proxy.RequestHandler {
	return m
}

func (*decisionHandlerRegistryMock) Writer() herodot.Writer {
	return nil
}

func (*decisionHandlerRegistryMock) Logger() *logrusx.Logger {
	return logrusx.New("", "")
}

func (m *decisionHandlerRegistryMock) Match(ctx context.Context, method string, u *url.URL, _ rule.Protocol) (*rule.Rule, error) {
	args := m.Called(ctx, method, u)
	return args.Get(0).(*rule.Rule), args.Error(1)
}

func (*decisionHandlerRegistryMock) HandleError(w http.ResponseWriter, r *http.Request, rl *rule.Rule, handleErr error) {
}

func (*decisionHandlerRegistryMock) HandleRequest(r *http.Request, rl *rule.Rule) (session *authn.AuthenticationSession, err error) {
	return &authn.AuthenticationSession{}, nil
}

func (*decisionHandlerRegistryMock) InitializeAuthnSession(r *http.Request, rl *rule.Rule) *authn.AuthenticationSession {
	return nil
}

func TestDecisionAPIHeaderUsage(t *testing.T) {
	r := new(decisionHandlerRegistryMock)
	h := api.NewJudgeHandler(r)
	defaultUrl := &url.URL{Scheme: "http", Host: "ory.sh", Path: "/foo"}
	defaultMethod := "GET"
	defaultTransform := func(req *http.Request) {}

	for _, tc := range []struct {
		name           string
		expectedMethod string
		expectedUrl    *url.URL
		transform      func(req *http.Request)
	}{
		{
			name:           "all arguments are taken from the url and request method",
			expectedUrl:    defaultUrl,
			expectedMethod: defaultMethod,
			transform:      defaultTransform,
		},
		{
			name:           "all arguments are taken from the url and request method, but scheme from URL TLS settings",
			expectedUrl:    &url.URL{Scheme: "https", Host: defaultUrl.Host, Path: defaultUrl.Path},
			expectedMethod: defaultMethod,
			transform: func(req *http.Request) {
				req.TLS = &tls.ConnectionState{}
			},
		},
		{
			name:           "all arguments are taken from the headers",
			expectedUrl:    &url.URL{Scheme: "https", Host: "test.dev", Path: "/bar"},
			expectedMethod: "POST",
			transform: func(req *http.Request) {
				req.Header.Add("X-Forwarded-Method", "POST")
				req.Header.Add("X-Forwarded-Proto", "https")
				req.Header.Add("X-Forwarded-Host", "test.dev")
				req.Header.Add("X-Forwarded-Uri", "/bar")
			},
		},
		{
			name:           "argument from the headers doesn't get url encoded",
			expectedUrl:    &url.URL{Scheme: "https", Host: "test.dev", Path: "/bar"},
			expectedMethod: "POST",
			transform: func(req *http.Request) {
				req.Header.Add("X-Forwarded-Method", "POST")
				req.Header.Add("X-Forwarded-Proto", "https")
				req.Header.Add("X-Forwarded-Host", "test.dev")
				req.Header.Add("X-Forwarded-Uri", "/bar?this=is&a=test")
			},
		},
		{
			name:           "only scheme is taken from the headers",
			expectedUrl:    &url.URL{Scheme: "https", Host: defaultUrl.Host, Path: defaultUrl.Path},
			expectedMethod: defaultMethod,
			transform: func(req *http.Request) {
				req.Header.Add("X-Forwarded-Proto", "https")
			},
		},
		{
			name:           "only method is taken from the headers",
			expectedUrl:    defaultUrl,
			expectedMethod: "POST",
			transform: func(req *http.Request) {
				req.Header.Add("X-Forwarded-Method", "POST")
			},
		},
		{
			name:           "only host is taken from the headers",
			expectedUrl:    &url.URL{Scheme: defaultUrl.Scheme, Host: "test.dev", Path: defaultUrl.Path},
			expectedMethod: defaultMethod,
			transform: func(req *http.Request) {
				req.Header.Add("X-Forwarded-Host", "test.dev")
			},
		},
		{
			name:           "only path is taken from the headers",
			expectedUrl:    &url.URL{Scheme: defaultUrl.Scheme, Host: defaultUrl.Host, Path: "/bar"},
			expectedMethod: defaultMethod,
			transform: func(req *http.Request) {
				req.Header.Add("X-Forwarded-Uri", "/bar")
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			reqUrl := *defaultUrl
			reqUrl.Path = api.DecisionPath + reqUrl.Path
			req := httptest.NewRequest(defaultMethod, reqUrl.String(), nil)
			tc.transform(req)

			r.On("Match", mock.Anything,
				mock.MatchedBy(func(val string) bool { return val == tc.expectedMethod }),
				mock.MatchedBy(func(val *url.URL) bool { return *val == *tc.expectedUrl })).
				Return(&rule.Rule{}, nil)
			h.ServeHTTP(res, req, nil)

			r.AssertExpectations(t)
		})
	}
}
