/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

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
	"github.com/ory/oathkeeper/rsakey"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	d := NewDirector(u, nil, nil, &rsakey.LocalManager{KeyStrength: 512})

	proxy := httptest.NewServer(&httputil.ReverseProxy{Director: d.Director, Transport: d})
	defer proxy.Close()

	publicRule := rule.Rule{MatchesMethods: []string{"GET"}, MatchesURLCompiled: mustCompileRegex(t, proxy.URL+"/users/[0-9]+"), Mode: rule.AnonymousMode}
	disabledRule := rule.Rule{MatchesMethods: []string{"GET"}, MatchesURLCompiled: mustCompileRegex(t, proxy.URL+"/users/[0-9]+"), Mode: rule.BypassMode}
	privateRule := rule.Rule{
		MatchesMethods:     []string{"GET"},
		MatchesURLCompiled: mustCompileRegex(t, proxy.URL+"/users/([0-9]+)"),
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
			d.Evaluator = evaluator.NewWardenEvaluator(nil, matcher, sdk, "")

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

func panicCompileRegex(pattern string) *regexp.Regexp {
	exp, err := regexp.Compile(pattern)
	if err != nil {
		panic(err.Error())
	}
	return exp
}

func BenchmarkDirector(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, r.Header.Get("Authorization"))
	}))
	defer backend.Close()

	logger := logrus.New()
	logger.Level = logrus.WarnLevel
	u, _ := url.Parse(backend.URL)
	d := NewDirector(u, nil, logger, &rsakey.LocalManager{KeyStrength: 512})

	proxy := httptest.NewServer(&httputil.ReverseProxy{Director: d.Director, Transport: d})
	defer proxy.Close()

	matcher := &rule.CachedMatcher{Rules: []rule.Rule{
		{MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(proxy.URL + "/users"), Mode: rule.AnonymousMode},
		{MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(proxy.URL + "/users/<[0-9]+>"), Mode: rule.AnonymousMode},
		{MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(proxy.URL + "/<[0-9]+>"), Mode: rule.AnonymousMode},
		{MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(proxy.URL + "/other/<.+>"), Mode: rule.AnonymousMode},
	}}
	d.Evaluator = evaluator.NewWardenEvaluator(logger, matcher, nil, "")

	req, _ := http.NewRequest("GET", proxy.URL+"/users", nil)

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
