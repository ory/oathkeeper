// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
	{
		ID:             "grpc1",
		Match:          &MatchGRPC{Authority: "<baz|bar>.example.com", FullMethod: "grpc.api/Call"},
		Description:    "gRPC Rule",
		Authorizer:     Handler{Handler: "allow", Config: []byte(`{"type":"any"}`)},
		Authenticators: []Handler{{Handler: "anonymous", Config: []byte(`{"name":"anonymous1"}`)}},
		Mutators:       []Handler{{Handler: "id_token", Config: []byte(`{"issuer":"anything"}`)}},
		Upstream:       Upstream{URL: "http://bar.example.com/", PreserveHost: false},
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
	{
		ID:             "grpc1",
		Match:          &MatchGRPC{Authority: "<{baz*,bar*}>.example.com", FullMethod: "grpc.api/Call"},
		Description:    "gRPC Rule",
		Authorizer:     Handler{Handler: "allow", Config: []byte(`{"type":"any"}`)},
		Authenticators: []Handler{{Handler: "anonymous", Config: []byte(`{"name":"anonymous1"}`)}},
		Mutators:       []Handler{{Handler: "id_token", Config: []byte(`{"issuer":"anything"}`)}},
		Upstream:       Upstream{URL: "http://bar.example.com/", PreserveHost: false},
	},
}

func TestMatcher(t *testing.T) {
	type m interface {
		Matcher
		Repository
	}

	var testMatcher = func(t *testing.T, matcher Matcher, method string, url string, protocol Protocol, expectErr bool, expect *Rule) {
		r, err := matcher.Match(context.Background(), method, mustParseURL(t, url), protocol)
		if expectErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			r.matchingEngine = nil
			assert.EqualValues(t, *expect, *r)
		}
	}

	for name, matcher := range map[string]m{
		"memory": NewRepositoryMemory(new(mockRepositoryRegistry)),
	} {
		t.Run(fmt.Sprintf("regexp matcher=%s", name), func(t *testing.T) {
			t.Run("case=empty", func(t *testing.T) {
				testMatcher(t, matcher, "GET", "https://localhost:34/baz", ProtocolHTTP, true, nil)
				testMatcher(t, matcher, "POST", "https://localhost:1234/foo", ProtocolHTTP, true, nil)
				testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", ProtocolHTTP, true, nil)
			})

			require.NoError(t, matcher.Set(context.Background(), testRules))

			t.Run("case=created", func(t *testing.T) {
				testMatcher(t, matcher, "GET", "https://localhost:34/baz", ProtocolHTTP, false, &testRules[1])
				testMatcher(t, matcher, "GET", "https://localhost:34/baz", ProtocolGRPC, true, nil)
				testMatcher(t, matcher, "POST", "https://localhost:1234/foo", ProtocolHTTP, false, &testRules[0])
				testMatcher(t, matcher, "POST", "https://localhost:1234/foo", ProtocolGRPC, true, nil)
				testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", ProtocolHTTP, true, nil)
				testMatcher(t, matcher, "POST", "grpc://bar.example.com/grpc.api/Call", ProtocolGRPC, false, &testRules[3])
			})

			t.Run("case=cache", func(t *testing.T) {
				r, err := matcher.Match(context.Background(), "GET", mustParseURL(t, "https://localhost:34/baz"), ProtocolHTTP)
				require.NoError(t, err)
				got, err := matcher.Get(context.Background(), r.ID)
				require.NoError(t, err)
				assert.NotEmpty(t, got.matchingEngine.Checksum())
			})

			t.Run("case=nil url", func(t *testing.T) {
				_, err := matcher.Match(context.Background(), "GET", nil, ProtocolHTTP)
				require.Error(t, err)
			})

			require.NoError(t, matcher.Set(context.Background(), testRules[1:]))

			t.Run("case=updated", func(t *testing.T) {
				testMatcher(t, matcher, "GET", "https://localhost:34/baz", ProtocolHTTP, false, &testRules[1])
				testMatcher(t, matcher, "POST", "https://localhost:1234/foo", ProtocolHTTP, true, nil)
				testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", ProtocolHTTP, true, nil)
			})
		})
		t.Run(fmt.Sprintf("glob matcher=%s", name), func(t *testing.T) {
			require.NoError(t, matcher.SetMatchingStrategy(context.Background(), configuration.Glob))
			require.NoError(t, matcher.Set(context.Background(), []Rule{}))
			t.Run("case=empty", func(t *testing.T) {
				testMatcher(t, matcher, "GET", "https://localhost:34/baz", ProtocolHTTP, true, nil)
				testMatcher(t, matcher, "POST", "https://localhost:1234/foo", ProtocolHTTP, true, nil)
				testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", ProtocolHTTP, true, nil)
			})

			require.NoError(t, matcher.Set(context.Background(), testRulesGlob))

			t.Run("case=created", func(t *testing.T) {
				testMatcher(t, matcher, "GET", "https://localhost:34/baz", ProtocolHTTP, false, &testRulesGlob[1])
				testMatcher(t, matcher, "POST", "https://localhost:1234/foo", ProtocolHTTP, false, &testRulesGlob[0])
				testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", ProtocolHTTP, true, nil)
				testMatcher(t, matcher, "POST", "grpc://bar.example.com/grpc.api/Call", ProtocolGRPC, false, &testRulesGlob[3])
			})

			t.Run("case=cache", func(t *testing.T) {
				r, err := matcher.Match(context.Background(), "GET", mustParseURL(t, "https://localhost:34/baz"), ProtocolHTTP)
				require.NoError(t, err)
				got, err := matcher.Get(context.Background(), r.ID)
				require.NoError(t, err)
				assert.NotEmpty(t, got.matchingEngine.Checksum())
			})

			require.NoError(t, matcher.Set(context.Background(), testRulesGlob[1:]))

			t.Run("case=updated", func(t *testing.T) {
				testMatcher(t, matcher, "GET", "https://localhost:34/baz", ProtocolHTTP, false, &testRulesGlob[1])
				testMatcher(t, matcher, "POST", "https://localhost:1234/foo", ProtocolHTTP, true, nil)
				testMatcher(t, matcher, "DELETE", "https://localhost:1234/foo", ProtocolHTTP, true, nil)
			})
		})
	}
}
