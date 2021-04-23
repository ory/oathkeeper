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
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/driver/configuration"
)

func mustParseURL(t *testing.T, u string) *url.URL {
	p, err := url.Parse(u)
	require.NoError(t, err)
	return p
}

var testRules = []Rule{
	{
		ID:             "foo1",
		Match:          &Match{URL: "https://localhost:1234/<foo|bar>", Methods: []string{"POST"}},
		Description:    "Create users rule",
		Authorizer:     Handler{Handler: "allow", Config: []byte(`{"type":"any"}`)},
		Authenticators: []Handler{{Handler: "anonymous", Config: []byte(`{"name":"anonymous1"}`)}},
		Mutators:       []Handler{{Handler: "id_token", Config: []byte(`{"issuer":"anything"}`)}},
		Upstream:       Upstream{URL: "http://localhost:1235/", StripPath: "/bar", PreserveHost: true},
	},
	{
		ID:             "foo2",
		Match:          &Match{URL: "https://localhost:34/<baz|bar>", Methods: []string{"GET"}},
		Description:    "Get users rule",
		Authorizer:     Handler{Handler: "deny", Config: []byte(`{"type":"any"}`)},
		Authenticators: []Handler{{Handler: "oauth2_introspection", Config: []byte(`{"name":"anonymous1"}`)}},
		Mutators:       []Handler{{Handler: "id_token", Config: []byte(`{"issuer":"anything"}`)}},
		Upstream:       Upstream{URL: "http://localhost:333/", StripPath: "/foo", PreserveHost: false},
	},
	{
		ID:             "foo3",
		Match:          &Match{URL: "https://localhost:343/<baz|bar>", Methods: []string{"GET"}},
		Description:    "Get users rule",
		Authorizer:     Handler{Handler: "deny"},
		Authenticators: []Handler{{Handler: "oauth2_introspection"}},
		Mutators:       []Handler{{Handler: "id_token"}},
		Upstream:       Upstream{URL: "http://localhost:3333/", StripPath: "/foo", PreserveHost: false},
	},
}

var testRulesGlob = []Rule{
	{
		ID:             "foo1",
		Match:          &Match{URL: "https://localhost:1234/<{foo*,bar*}>", Methods: []string{"POST"}},
		Description:    "Create users rule",
		Authorizer:     Handler{Handler: "allow", Config: []byte(`{"type":"any"}`)},
		Authenticators: []Handler{{Handler: "anonymous", Config: []byte(`{"name":"anonymous1"}`)}},
		Mutators:       []Handler{{Handler: "id_token", Config: []byte(`{"issuer":"anything"}`)}},
		Upstream:       Upstream{URL: "http://localhost:1235/", StripPath: "/bar", PreserveHost: true},
	},
	{
		ID:             "foo2",
		Match:          &Match{URL: "https://localhost:34/<{baz*,bar*}>", Methods: []string{"GET"}},
		Description:    "Get users rule",
		Authorizer:     Handler{Handler: "deny", Config: []byte(`{"type":"any"}`)},
		Authenticators: []Handler{{Handler: "oauth2_introspection", Config: []byte(`{"name":"anonymous1"}`)}},
		Mutators:       []Handler{{Handler: "id_token", Config: []byte(`{"issuer":"anything"}`)}},
		Upstream:       Upstream{URL: "http://localhost:333/", StripPath: "/foo", PreserveHost: false},
	},
	{
		ID:             "foo3",
		Match:          &Match{URL: "https://localhost:343/<{baz*,bar*}>", Methods: []string{"GET"}},
		Description:    "Get users rule",
		Authorizer:     Handler{Handler: "deny"},
		Authenticators: []Handler{{Handler: "oauth2_introspection"}},
		Mutators:       []Handler{{Handler: "id_token"}},
		Upstream:       Upstream{URL: "http://localhost:3333/", StripPath: "/foo", PreserveHost: false},
	},
}

func TestMatcher(t *testing.T) {
	type m interface {
		Matcher
		Repository
	}

	var testMatcher = func(t *testing.T, matcher Matcher, method string, url string, expectErr bool, expect *Rule) {
		r, err := matcher.Match(context.Background(), method, mustParseURL(t, url))
		if expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.EqualValues(t, *expect, *r)
		}
	}

	for name, matcher := range map[string]m{
		"memory": NewRepositoryMemory(new(mockRepositoryRegistry), new(mockHealthEventManager)),
	} {
		t.Run(fmt.Sprintf("regexp matcher=%s", name), func(t *testing.T) {
			t.Run("case=empty", func(t *testing.T) {
				testMatcher(t, matcher, "GET", "https://localhost:34/baz", true, nil)
				testMatcher(t, matcher, "POST", "https://localhost:1234/foo", true, nil)
				testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", true, nil)
			})

			require.NoError(t, matcher.Set(context.Background(), testRules))

			t.Run("case=created", func(t *testing.T) {
				testMatcher(t, matcher, "GET", "https://localhost:34/baz", false, &testRules[1])
				testMatcher(t, matcher, "POST", "https://localhost:1234/foo", false, &testRules[0])
				testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", true, nil)
			})

			t.Run("case=cache", func(t *testing.T) {
				r, err := matcher.Match(context.Background(), "GET", mustParseURL(t, "https://localhost:34/baz"))
				require.NoError(t, err)
				got, err := matcher.Get(context.Background(), r.ID)
				require.NoError(t, err)
				assert.NotEmpty(t, got.matchingEngine.Checksum())
			})

			t.Run("case=nil url", func(t *testing.T) {
				_, err := matcher.Match(context.Background(), "GET", nil)
				require.Error(t, err)
			})

			require.NoError(t, matcher.Set(context.Background(), testRules[1:]))

			t.Run("case=updated", func(t *testing.T) {
				testMatcher(t, matcher, "GET", "https://localhost:34/baz", false, &testRules[1])
				testMatcher(t, matcher, "POST", "https://localhost:1234/foo", true, nil)
				testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", true, nil)
			})
		})
		t.Run(fmt.Sprintf("glob matcher=%s", name), func(t *testing.T) {
			require.NoError(t, matcher.SetMatchingStrategy(context.Background(), configuration.Glob))
			require.NoError(t, matcher.Set(context.Background(), []Rule{}))
			t.Run("case=empty", func(t *testing.T) {
				testMatcher(t, matcher, "GET", "https://localhost:34/baz", true, nil)
				testMatcher(t, matcher, "POST", "https://localhost:1234/foo", true, nil)
				testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", true, nil)
			})

			require.NoError(t, matcher.Set(context.Background(), testRulesGlob))

			t.Run("case=created", func(t *testing.T) {
				testMatcher(t, matcher, "GET", "https://localhost:34/baz", false, &testRulesGlob[1])
				testMatcher(t, matcher, "POST", "https://localhost:1234/foo", false, &testRulesGlob[0])
				testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", true, nil)
			})

			t.Run("case=cache", func(t *testing.T) {
				r, err := matcher.Match(context.Background(), "GET", mustParseURL(t, "https://localhost:34/baz"))
				require.NoError(t, err)
				got, err := matcher.Get(context.Background(), r.ID)
				require.NoError(t, err)
				assert.NotEmpty(t, got.matchingEngine.Checksum())
			})

			require.NoError(t, matcher.Set(context.Background(), testRulesGlob[1:]))

			t.Run("case=updated", func(t *testing.T) {
				testMatcher(t, matcher, "GET", "https://localhost:34/baz", false, &testRulesGlob[1])
				testMatcher(t, matcher, "POST", "https://localhost:1234/foo", true, nil)
				testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", true, nil)
			})
		})
	}
}
