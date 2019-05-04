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

package judge

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rule"
)

func TestJudge(t *testing.T) {
	matcher := &rule.CachedMatcher{Rules: map[string]rule.Rule{}}
	rh := proxy.NewRequestHandler(
		nil,
		[]proxy.Authenticator{proxy.NewAuthenticatorNoOp(), proxy.NewAuthenticatorAnonymous("anonymous"), proxy.NewAuthenticatorBroken()},
		[]proxy.Authorizer{proxy.NewAuthorizerAllow(), proxy.NewAuthorizerDeny()},
		[]proxy.CredentialsIssuer{proxy.NewCredentialsIssuerNoOp(), proxy.NewCredentialsIssuerBroken()},
	)

	router := httprouter.New()
	d := NewHandler(rh, nil, matcher, router)

	ts := httptest.NewServer(d)
	defer ts.Close()

	ruleNoOpAuthenticator := rule.Rule{
		Match:             rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-noop/<[0-9]+>"},
		Authenticators:    []rule.RuleHandler{{Handler: "noop"}},
		Authorizer:        rule.RuleHandler{Handler: proxy.NewAuthorizerAllow().GetID()},
		CredentialsIssuer: rule.RuleHandler{Handler: proxy.NewCredentialsIssuerNoOp().GetID()},
		Upstream:          rule.Upstream{URL: ""},
	}
	ruleNoOpAuthenticatorModifyUpstream := rule.Rule{
		Match:             rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/strip-path/authn-noop/<[0-9]+>"},
		Authenticators:    []rule.RuleHandler{{Handler: "noop"}},
		Authorizer:        rule.RuleHandler{Handler: proxy.NewAuthorizerAllow().GetID()},
		CredentialsIssuer: rule.RuleHandler{Handler: proxy.NewCredentialsIssuerNoOp().GetID()},
		Upstream:          rule.Upstream{URL: "", StripPath: "/strip-path/", PreserveHost: true},
	}

	for k, tc := range []struct {
		url       string
		code      int
		messages  []string
		rules     []rule.Rule
		transform func(r *http.Request)
		authz     string
		d         string
	}{
		{
			d:     "should fail because url does not exist in rule set",
			url:   ts.URL + "/judge" + "/invalid",
			rules: []rule.Rule{},
			code:  http.StatusNotFound,
		},
		{
			d:     "should fail because url does exist but is matched by two rules",
			url:   ts.URL + "/judge" + "/authn-noop/1234",
			rules: []rule.Rule{ruleNoOpAuthenticator, ruleNoOpAuthenticator},
			code:  http.StatusInternalServerError,
		},
		{
			d:     "should pass",
			url:   ts.URL + "/judge" + "/authn-noop/1234",
			rules: []rule.Rule{ruleNoOpAuthenticator},
			code:  http.StatusOK,
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
			},
			authz: "bearer token",
		},
		{
			d:     "should pass",
			url:   ts.URL + "/judge" + "/strip-path/authn-noop/1234",
			rules: []rule.Rule{ruleNoOpAuthenticatorModifyUpstream},
			code:  http.StatusOK,
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
			},
			authz: "bearer token",
		},
		{
			d:   "should fail because no authorizer was configured",
			url: ts.URL + "/judge" + "/authn-anon/authz-none/cred-none/1234",
			rules: []rule.Rule{{
				Match:          rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-none/cred-none/<[0-9]+>"},
				Authenticators: []rule.RuleHandler{{Handler: "anonymous"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
			},
			code: http.StatusUnauthorized,
		},
		{
			d:   "should fail because no credentials issuer was configured",
			url: ts.URL + "/judge" + "/authn-anon/authz-allow/cred-none/1234",
			rules: []rule.Rule{{
				Match:          rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-allow/cred-none/<[0-9]+>"},
				Authenticators: []rule.RuleHandler{{Handler: "anonymous"}},
				Authorizer:     rule.RuleHandler{Handler: "allow"},
				Upstream:       rule.Upstream{URL: ""},
			}},
			code: http.StatusInternalServerError,
		},
		{
			d:   "should pass with anonymous and everything else set to noop",
			url: ts.URL + "/judge" + "/authn-anon/authz-allow/cred-noop/1234",
			rules: []rule.Rule{{
				Match:             rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-allow/cred-noop/<[0-9]+>"},
				Authenticators:    []rule.RuleHandler{{Handler: "anonymous"}},
				Authorizer:        rule.RuleHandler{Handler: "allow"},
				CredentialsIssuer: rule.RuleHandler{Handler: "noop"},
				Upstream:          rule.Upstream{URL: ""},
			}},
			code:  http.StatusOK,
			authz: "",
		},
		{
			d:   "should fail when authorizer fails",
			url: ts.URL + "/judge" + "/authn-anon/authz-deny/cred-noop/1234",
			rules: []rule.Rule{{
				Match:             rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-deny/cred-noop/<[0-9]+>"},
				Authenticators:    []rule.RuleHandler{{Handler: "anonymous"}},
				Authorizer:        rule.RuleHandler{Handler: "deny"},
				CredentialsIssuer: rule.RuleHandler{Handler: "noop"},
				Upstream:          rule.Upstream{URL: ""},
			}},
			code: http.StatusForbidden,
		},
		{
			d:   "should fail when authenticator fails",
			url: ts.URL + "/judge" + "/authn-broken/authz-none/cred-none/1234",
			rules: []rule.Rule{{
				Match:          rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-broken/authz-none/cred-none/<[0-9]+>"},
				Authenticators: []rule.RuleHandler{{Handler: "broken"}},
				Upstream:       rule.Upstream{URL: ""},
			}},
			code: http.StatusUnauthorized,
		},
		{
			d:   "should fail when credentials issuer fails",
			url: ts.URL + "/judge" + "/authn-anonymous/authz-allow/cred-broken/1234",
			rules: []rule.Rule{{
				Match:             rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-anonymous/authz-allow/cred-broken/<[0-9]+>"},
				Authenticators:    []rule.RuleHandler{{Handler: "anonymous"}},
				Authorizer:        rule.RuleHandler{Handler: "allow"},
				CredentialsIssuer: rule.RuleHandler{Handler: "broken"},
				Upstream:          rule.Upstream{URL: ""},
			}},
			code: http.StatusInternalServerError,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			matcher.Rules = map[string]rule.Rule{}
			for k, r := range tc.rules {
				matcher.Rules[strconv.Itoa(k)] = r
			}

			req, err := http.NewRequest("GET", tc.url, nil)
			require.NoError(t, err)
			if tc.transform != nil {
				tc.transform(req)
			}

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tc.authz, res.Header.Get("Authorization"))
			assert.Equal(t, tc.code, res.StatusCode)
		})
	}
}
