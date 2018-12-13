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

package rule

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/sdk/go/oathkeeper"
)

var testRules = []Rule{
	{
		ID: "foo1",
		Match: RuleMatch{
			URL:     "https://localhost:1234/<foo|bar>",
			Methods: []string{"POST"},
		},
		Description:       "Create users rule",
		Authorizer:        RuleHandler{Handler: "allow", Config: []byte(`{"type":"any"}`)},
		Authenticators:    []RuleHandler{{Handler: "anonymous", Config: []byte(`{"name":"anonymous1"}`)}},
		CredentialsIssuer: RuleHandler{Handler: "id_token", Config: []byte(`{"issuer":"anything"}`)},
		Upstream: Upstream{
			URL:          "http://localhost:1235/",
			StripPath:    "/bar",
			PreserveHost: true,
		},
	},
	{
		ID: "foo2",
		Match: RuleMatch{
			URL:     "https://localhost:34/<baz|bar>",
			Methods: []string{"GET"},
		},
		Description:       "Get users rule",
		Authorizer:        RuleHandler{Handler: "deny", Config: []byte(`{"type":"any"}`)},
		Authenticators:    []RuleHandler{{Handler: "oauth2_introspection", Config: []byte(`{"name":"anonymous1"}`)}},
		CredentialsIssuer: RuleHandler{Handler: "id_token", Config: []byte(`{"issuer":"anything"}`)},
		Upstream: Upstream{
			URL:          "http://localhost:333/",
			StripPath:    "/foo",
			PreserveHost: false,
		},
	},
	{
		ID: "foo3",
		Match: RuleMatch{
			URL:     "https://localhost:343/<baz|bar>",
			Methods: []string{"GET"},
		},
		Description:       "Get users rule",
		Authorizer:        RuleHandler{Handler: "deny"},
		Authenticators:    []RuleHandler{{Handler: "oauth2_introspection"}},
		CredentialsIssuer: RuleHandler{Handler: "id_token"},
		Upstream: Upstream{
			URL:          "http://localhost:3333/",
			StripPath:    "/foo",
			PreserveHost: false,
		},
	},
}

func mustParseURL(t *testing.T, u string) *url.URL {
	p, err := url.Parse(u)
	require.NoError(t, err)
	return p
}

func TestMatcher(t *testing.T) {
	manager := NewMemoryManager()
	handler := &Handler{
		H: herodot.NewJSONWriter(logrus.New()),
		M: manager,
	}
	router := httprouter.New()
	handler.SetRoutes(router)
	server := httptest.NewServer(router)

	matchers := map[string]Matcher{
		"memory": NewCachedMatcher(manager),
		"http":   NewHTTPMatcher(oathkeeper.NewSDK(server.URL)),
	}

	var testMatcher = func(t *testing.T, matcher Matcher, method string, url string, expectErr bool, expect *Rule) {
		r, err := matcher.MatchRule(method, mustParseURL(t, url))
		if expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.EqualValues(t, *expect, *r)
		}
	}

	for name, matcher := range matchers {
		t.Run("matcher="+name+"/case=empty", func(t *testing.T) {
			require.NoError(t, matcher.Refresh())
			testMatcher(t, matcher, "GET", "https://localhost:34/baz", true, nil)
			testMatcher(t, matcher, "POST", "https://localhost:1234/foo", true, nil)
			testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", true, nil)
		})
	}

	for _, tr := range testRules {
		require.NoError(t, manager.CreateRule(&tr))
	}

	for name, matcher := range matchers {
		t.Run("matcher="+name+"/case=created", func(t *testing.T) {
			require.NoError(t, matcher.Refresh())
			testMatcher(t, matcher, "GET", "https://localhost:34/baz", false, &testRules[1])
			testMatcher(t, matcher, "POST", "https://localhost:1234/foo", false, &testRules[0])
			testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", true, nil)
		})
	}

	require.NoError(t, manager.DeleteRule(testRules[0].ID))

	for name, matcher := range matchers {
		t.Run("matcher="+name+"/case=updated", func(t *testing.T) {
			require.NoError(t, matcher.Refresh())
			testMatcher(t, matcher, "GET", "https://localhost:34/baz", false, &testRules[1])
			testMatcher(t, matcher, "POST", "https://localhost:1234/foo", true, nil)
			testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", true, nil)
		})
	}
}
