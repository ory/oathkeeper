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

package api_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ory/viper"

	"github.com/urfave/negroni"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/rule"
)

func TestDecisionAPI(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	viper.Set(configuration.ViperKeyAuthenticatorNoopIsEnabled, true)
	viper.Set(configuration.ViperKeyAuthenticatorUnauthorizedIsEnabled, true)
	viper.Set(configuration.ViperKeyAuthenticatorAnonymousIsEnabled, true)
	viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, true)
	viper.Set(configuration.ViperKeyAuthorizerDenyIsEnabled, true)
	viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, true)
	viper.Set(configuration.ViperKeyErrorsWWWAuthenticateIsEnabled, true)
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
		messages    []string
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

				entireBody, err := ioutil.ReadAll(res.Body)
				require.NoError(t, err)
				defer res.Body.Close()

				assert.Equal(t, tc.authz, res.Header.Get("Authorization"))
				assert.Equal(t, tc.code, res.StatusCode)
				assert.Equal(t, strconv.Itoa(len(entireBody)), res.Header.Get("Content-Length"))
			}
			t.Run("regexp", func(t *testing.T) {
				reg.RuleRepository().(*rule.RepositoryMemory).WithRules(tc.rulesRegexp)
				testFunc(configuration.Regexp)
			})
			t.Run("glob", func(t *testing.T) {
				reg.RuleRepository().(*rule.RepositoryMemory).WithRules(tc.rulesRegexp)
				testFunc(configuration.Glob)
			})
		})
	}
}
