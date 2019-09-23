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

package proxy_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"

	"github.com/ory/viper"

	"github.com/ory/x/urlx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/proxy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/rule"
)

// type jurorDenyAll struct{}
//
// func (j *jurorDenyAll) GetID() string {
//	return "pass_through_deny"
// }
//
// func (j jurorDenyAll) Try(r *http.Request, rl *rule.Rule, u *url.URL) (*Session, error) {
//	return nil, errors.WithStack(helper.ErrUnauthorized)
// }
//
// type jurorAcceptAll struct{}
//
// func (j *jurorAcceptAll) GetID() string {
//	return "pass_through_accept"
// }
//
// func (j jurorAcceptAll) Try(r *http.Request, rl *rule.Rule, u *url.URL) (*Session, error) {
//	return &Session{
//		Subject:   "",
//		Anonymous: true,
//		ClientID:  "",
//		Disabled:  true,
//	}, nil
// }

func TestProxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// assert.NotEmpty(t, helper.BearerTokenFromRequest(r))
		fmt.Fprint(w, "authorization="+r.Header.Get("Authorization")+"\n")
		fmt.Fprint(w, "host="+r.Host+"\n")
		fmt.Fprint(w, "url="+r.URL.String())
	}))
	defer backend.Close()

	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf).WithBrokenPipelineMutator()

	d := reg.Proxy()
	ts := httptest.NewServer(&httputil.ReverseProxy{Director: d.Director, Transport: d})
	defer ts.Close()

	viper.Set(configuration.ViperKeyAuthenticatorNoopIsEnabled, true)
	viper.Set(configuration.ViperKeyAuthenticatorUnauthorizedIsEnabled, true)
	viper.Set(configuration.ViperKeyAuthenticatorAnonymousIsEnabled, true)
	viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, true)
	viper.Set(configuration.ViperKeyAuthorizerDenyIsEnabled, true)
	viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, true)

	ruleNoOpAuthenticator := rule.Rule{
		Match:          rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-noop/<[0-9]+>"},
		Authenticators: []rule.RuleHandler{{Handler: "noop"}},
		Authorizer:     rule.RuleHandler{Handler: "allow"},
		Mutators:       []rule.RuleHandler{{Handler: "noop"}},
		Upstream:       rule.Upstream{URL: backend.URL},
	}
	ruleNoOpAuthenticatorModifyUpstream := rule.Rule{
		Match:          rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/strip-path/authn-noop/<[0-9]+>"},
		Authenticators: []rule.RuleHandler{{Handler: "noop"}},
		Authorizer:     rule.RuleHandler{Handler: "allow"},
		Mutators:       []rule.RuleHandler{{Handler: "noop"}},
		Upstream:       rule.Upstream{URL: backend.URL, StripPath: "/strip-path/", PreserveHost: true},
	}

	// acceptRuleStripHost := rule.Rule{MatchesMethods: []string{"GET"}, MatchesURLCompiled: mustCompileRegex(t, proxy.URL+"/users/<[0-9]+>"), Mode: "pass_through_accept", Upstream: rule.Upstream{URLParsed: u, StripPath: "/users/", PreserveHost: true}}
	// acceptRuleStripHostWithoutTrailing := rule.Rule{MatchesMethods: []string{"GET"}, MatchesURLCompiled: mustCompileRegex(t, proxy.URL+"/users/<[0-9]+>"), Mode: "pass_through_accept", Upstream: rule.Upstream{URLParsed: u, StripPath: "/users", PreserveHost: true}}
	// acceptRuleStripHostWithoutTrailing2 := rule.Rule{MatchesMethods: []string{"GET"}, MatchesURLCompiled: mustCompileRegex(t, proxy.URL+"/users/<[0-9]+>"), Mode: "pass_through_accept", Upstream: rule.Upstream{URLParsed: u, StripPath: "users", PreserveHost: true}}
	// denyRule := rule.Rule{MatchesMethods: []string{"GET"}, MatchesURLCompiled: mustCompileRegex(t, proxy.URL+"/users/<[0-9]+>"), Mode: "pass_through_deny", Upstream: rule.Upstream{URLParsed: u}}

	for k, tc := range []struct {
		url       string
		code      int
		messages  []string
		rules     []rule.Rule
		transform func(r *http.Request)
		d         string
	}{
		{
			d:     "should fail because url does not exist in rule set",
			url:   ts.URL + "/invalid",
			rules: []rule.Rule{},
			code:  http.StatusNotFound,
		},
		{
			d:     "should fail because url does exist but is matched by two rules",
			url:   ts.URL + "/authn-noop/1234",
			rules: []rule.Rule{ruleNoOpAuthenticator, ruleNoOpAuthenticator},
			code:  http.StatusInternalServerError,
		},
		{
			d:     "should pass",
			url:   ts.URL + "/authn-noop/1234",
			rules: []rule.Rule{ruleNoOpAuthenticator},
			code:  http.StatusOK,
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
			},
			messages: []string{
				"authorization=bearer token",
				"url=/authn-noop/1234",
				"host=" + urlx.ParseOrPanic(backend.URL).Host,
			},
		},
		{
			d:     "should pass",
			url:   ts.URL + "/strip-path/authn-noop/1234",
			rules: []rule.Rule{ruleNoOpAuthenticatorModifyUpstream},
			code:  http.StatusOK,
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
			},
			messages: []string{
				"authorization=bearer token",
				"url=/authn-noop/1234",
				"host=" + urlx.ParseOrPanic(ts.URL).Host,
			},
		},
		{
			d:   "should fail because no authorizer was configured",
			url: ts.URL + "/authn-anon/authz-none/cred-none/1234",
			rules: []rule.Rule{{
				Match:          rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-none/cred-none/<[0-9]+>"},
				Authenticators: []rule.RuleHandler{{Handler: "anonymous"}},
				Upstream:       rule.Upstream{URL: backend.URL},
			}},
			transform: func(r *http.Request) {
				r.Header.Add("Authorization", "bearer token")
			},
			code: http.StatusUnauthorized,
		},
		{
			d:   "should fail because no credentials issuer was configured",
			url: ts.URL + "/authn-anon/authz-allow/cred-none/1234",
			rules: []rule.Rule{{
				Match:          rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-allow/cred-none/<[0-9]+>"},
				Authenticators: []rule.RuleHandler{{Handler: "anonymous"}},
				Authorizer:     rule.RuleHandler{Handler: "allow"},
				Upstream:       rule.Upstream{URL: backend.URL},
			}},
			code: http.StatusInternalServerError,
		},
		{
			d:   "should pass with anonymous and everything else set to noop",
			url: ts.URL + "/authn-anon/authz-allow/cred-noop/1234",
			rules: []rule.Rule{{
				Match:          rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-allow/cred-noop/<[0-9]+>"},
				Authenticators: []rule.RuleHandler{{Handler: "anonymous"}},
				Authorizer:     rule.RuleHandler{Handler: "allow"},
				Mutators:       []rule.RuleHandler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: backend.URL},
			}},
			code: http.StatusOK,
			messages: []string{
				"authorization=",
				"url=/authn-anon/authz-allow/cred-noop/1234",
				"host=" + urlx.ParseOrPanic(backend.URL).Host,
			},
		},
		{
			d:   "should fail when authorizer fails",
			url: ts.URL + "/authn-anon/authz-deny/cred-noop/1234",
			rules: []rule.Rule{{
				Match:          rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-deny/cred-noop/<[0-9]+>"},
				Authenticators: []rule.RuleHandler{{Handler: "anonymous"}},
				Authorizer:     rule.RuleHandler{Handler: "deny"},
				Mutators:       []rule.RuleHandler{{Handler: "noop"}},
				Upstream:       rule.Upstream{URL: backend.URL},
			}},
			code: http.StatusForbidden,
		},
		{
			d:   "should fail when authenticator fails",
			url: ts.URL + "/authn-broken/authz-none/cred-none/1234",
			rules: []rule.Rule{{
				Match:          rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-broken/authz-none/cred-none/<[0-9]+>"},
				Authenticators: []rule.RuleHandler{{Handler: "unauthorized"}},
				Upstream:       rule.Upstream{URL: backend.URL},
			}},
			code: http.StatusUnauthorized,
		},
		{
			d:   "should fail because no mutator was configured",
			url: ts.URL + "/authn-anon/authz-deny/cred-noop/1234",
			rules: []rule.Rule{{
				Match:          rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-deny/cred-noop/<[0-9]+>"},
				Authenticators: []rule.RuleHandler{{Handler: "anonymous"}},
				Authorizer:     rule.RuleHandler{Handler: "allow"},
				Upstream:       rule.Upstream{URL: backend.URL},
			}},
			code: http.StatusInternalServerError,
		},
		{
			d:   "should fail when one of the mutators fails",
			url: ts.URL + "/authn-anon/authz-deny/cred-noop/1234",
			rules: []rule.Rule{{
				Match:          rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-anon/authz-deny/cred-noop/<[0-9]+>"},
				Authenticators: []rule.RuleHandler{{Handler: "anonymous"}},
				Authorizer:     rule.RuleHandler{Handler: "allow"},
				Mutators:       []rule.RuleHandler{{Handler: "noop"}, {Handler: "broken"}},
				Upstream:       rule.Upstream{URL: backend.URL},
			}},
			code: http.StatusInternalServerError,
		},
		{
			d:   "should fail when credentials issuer fails",
			url: ts.URL + "/authn-anonymous/authz-allow/cred-broken/1234",
			rules: []rule.Rule{{
				Match:          rule.RuleMatch{Methods: []string{"GET"}, URL: ts.URL + "/authn-anonymous/authz-allow/cred-broken/<[0-9]+>"},
				Authenticators: []rule.RuleHandler{{Handler: "anonymous"}},
				Authorizer:     rule.RuleHandler{Handler: "allow"},
				Mutators:       []rule.RuleHandler{{Handler: "broken"}},
				Upstream:       rule.Upstream{URL: backend.URL},
			}},
			code: http.StatusInternalServerError,
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
			reg.RuleRepository().(*rule.RepositoryMemory).WithRules(tc.rules)

			req, err := http.NewRequest("GET", tc.url, nil)
			require.NoError(t, err)
			if tc.transform != nil {
				tc.transform(req)
			}

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			greeting, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)

			assert.Equal(t, tc.code, res.StatusCode, "%s", res.Body)
			for _, m := range tc.messages {
				assert.True(t, strings.Contains(string(greeting), m), `Value "%s" not found in message:
%s
proxy_url=%s
backend_url=%s
`, m, greeting, ts.URL, backend.URL)
			}
		})
	}
}

func TestConfigureBackendURL(t *testing.T) {
	for k, tc := range []struct {
		r     *http.Request
		rl    *rule.Rule
		eURL  string
		eHost string
	}{
		{
			r:     &http.Request{Host: "localhost", URL: &url.URL{Path: "/api/users/1234", Scheme: "http"}},
			rl:    &rule.Rule{Upstream: rule.Upstream{URL: "http://localhost/"}},
			eURL:  "http://localhost/api/users/1234",
			eHost: "localhost",
		},
		{
			r:     &http.Request{Host: "localhost:3000", URL: &url.URL{Path: "/api/users/1234", Scheme: "http"}},
			rl:    &rule.Rule{Upstream: rule.Upstream{URL: "http://localhost:4000/", PreserveHost: true}},
			eURL:  "http://localhost:4000/api/users/1234",
			eHost: "localhost:3000",
		},
		{
			r:     &http.Request{Host: "localhost:3000", URL: &url.URL{Path: "/api/users/1234", Scheme: "http"}},
			rl:    &rule.Rule{Upstream: rule.Upstream{URL: "http://localhost:4000/", PreserveHost: true, StripPath: "/api/"}},
			eURL:  "http://localhost:4000/users/1234",
			eHost: "localhost:3000",
		},
		{
			r:     &http.Request{Host: "localhost:3000", URL: &url.URL{Path: "/api/users/1234", Scheme: "http"}},
			rl:    &rule.Rule{Upstream: rule.Upstream{URL: "http://localhost:4000/", PreserveHost: true, StripPath: "api/"}},
			eURL:  "http://localhost:4000/users/1234",
			eHost: "localhost:3000",
		},
		{
			r:     &http.Request{Host: "localhost:3000", URL: &url.URL{Path: "/api/users/1234", Scheme: "http"}},
			rl:    &rule.Rule{Upstream: rule.Upstream{URL: "http://localhost:4000/", PreserveHost: true, StripPath: "/api"}},
			eURL:  "http://localhost:4000/users/1234",
			eHost: "localhost:3000",
		},
		{
			r:     &http.Request{Host: "localhost:3000", URL: &url.URL{Path: "/api/users/1234", Scheme: "http"}},
			rl:    &rule.Rule{Upstream: rule.Upstream{URL: "http://localhost:4000/", PreserveHost: true, StripPath: "api"}},
			eURL:  "http://localhost:4000/users/1234",
			eHost: "localhost:3000",
		},
		{
			r:     &http.Request{Host: "localhost:3000", URL: &url.URL{Path: "/api/users/1234", Scheme: "http"}},
			rl:    &rule.Rule{Upstream: rule.Upstream{URL: "http://localhost:4000/foo/", PreserveHost: true, StripPath: "api"}},
			eURL:  "http://localhost:4000/foo/users/1234",
			eHost: "localhost:3000",
		},
		{
			r:     &http.Request{Host: "localhost:3000", URL: &url.URL{Path: "/api/users/1234", Scheme: "http"}},
			rl:    &rule.Rule{Upstream: rule.Upstream{URL: "http://localhost:4000/foo/", PreserveHost: true, StripPath: "api"}},
			eURL:  "http://localhost:4000/foo/users/1234",
			eHost: "localhost:3000",
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			require.NoError(t, proxy.ConfigureBackendURL(tc.r, tc.rl))
			assert.EqualValues(t, tc.eURL, tc.r.URL.String())
			assert.EqualValues(t, tc.eHost, tc.r.Host)
		})
	}
}

// func panicCompileRegex(pattern string) *regexp.Regexp {
//	exp, err := regexp.Compile(pattern)
//	if err != nil {
//		panic(err.Error())
//	}
//	return exp
// }

//
// func BenchmarkDirector(b *testing.B) {
//	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		fmt.Fprint(w, "authorization="+r.Header.Get("Authorization"))
//		fmt.Fprint(w, "host="+r.Header.Get("Host"))
//		fmt.Fprint(w, "url="+r.URL.String())
//		fmt.Fprint(w, "path="+r.URL.Path)
//	}))
//	defer backend.Close()
//
//	logger := logrus.New()
//	logger.Level = logrus.WarnLevel
//	u, _ := url.Parse(backend.URL)
//	d := NewProxy(nil, logger, &rsakey.LocalManager{KeyStrength: 512})
//
//	p := httptest.NewServer(&httputil.ReverseProxy{Director: d.Director, Transport: d})
//	defer p.Close()
//
//	jt := &JurorPassThrough{L: logrus.New()}
//	matcher := &rule.CachedMatcher{Rules: map[string]rule.Rule{
//		"A": {MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(p.URL + "/users"), Mode: jt.GetID(), Upstream: rule.Upstream{URLParsed: u}},
//		"B": {MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(p.URL + "/users/<[0-9]+>"), Mode: jt.GetID(), Upstream: rule.Upstream{URLParsed: u}},
//		"C": {MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(p.URL + "/<[0-9]+>"), Mode: jt.GetID(), Upstream: rule.Upstream{URLParsed: u}},
//		"D": {MatchesMethods: []string{"GET"}, MatchesURLCompiled: panicCompileRegex(p.URL + "/other/<.+>"), Mode: jt.GetID(), Upstream: rule.Upstream{URLParsed: u}},
//	}}
//	d.Judge = NewRequestHandler(logger, matcher, "", []Juror{jt})
//
//	req, _ := http.NewRequest("GET", p.URL+"/users", nil)
//
//	b.Run("case=fetch_user_endpoint", func(b *testing.B) {
//		for n := 0; n < b.N; n++ {
//			res, err := http.DefaultClient.Do(req)
//			if err != nil {
//				b.FailNow()
//			}
//
//			if res.StatusCode != http.StatusOK {
//				b.FailNow()
//			}
//		}
//	})
// }
