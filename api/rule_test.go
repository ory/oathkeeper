// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/x"

	"github.com/ory/x/pointerx"

	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/internal/httpclient/client"
	sdkrule "github.com/ory/oathkeeper/internal/httpclient/client/api"
	"github.com/ory/oathkeeper/rule"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf).WithBrokenPipelineMutator()

	router := x.NewAPIRouter()
	reg.RuleHandler().SetRoutes(router)
	server := httptest.NewServer(router)

	u, err := url.ParseRequestURI(server.URL)
	require.NoError(t, err)

	cl := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:     u.Host,
		BasePath: u.Path,
		Schemes:  []string{u.Scheme},
	})

	rulesRegexp := []rule.Rule{
		{
			ID: "foo1",
			Match: &rule.Match{
				URL:     "https://localhost:1234/<foo|bar>",
				Methods: []string{"POST"},
			},
			Description:    "Create users rule",
			Authorizer:     rule.Handler{Handler: "allow", Config: json.RawMessage(`{"type":"any"}`)},
			Authenticators: []rule.Handler{{Handler: "anonymous", Config: json.RawMessage(`{"name":"anonymous1"}`)}},
			Mutators:       []rule.Handler{{Handler: "id_token", Config: json.RawMessage(`{"issuer":"anything"}`)}},
			Upstream: rule.Upstream{
				URL:          "http://localhost:1235/",
				StripPath:    "/bar",
				PreserveHost: true,
			},
		},
		{
			ID: "foo2",
			Match: &rule.Match{
				URL:     "https://localhost:34/<baz|bar>",
				Methods: []string{"GET"},
			},
			Description:    "Get users rule",
			Authorizer:     rule.Handler{Handler: "deny", Config: json.RawMessage(`{"type":"any"}`)},
			Authenticators: []rule.Handler{{Handler: "oauth2_introspection", Config: json.RawMessage(`{"name":"anonymous1"}`)}},
			Mutators:       []rule.Handler{{Handler: "id_token", Config: json.RawMessage(`{"issuer":"anything"}`)}, {Handler: "headers", Config: json.RawMessage(`{"headers":{"X-User":"user"}}`)}},
			Upstream: rule.Upstream{
				URL:          "http://localhost:333/",
				StripPath:    "/foo",
				PreserveHost: false,
			},
		},
	}
	rulesGlob := []rule.Rule{
		{
			ID: "foo1",
			Match: &rule.Match{
				URL:     "https://localhost:1234/<{foo*,bar*}>>",
				Methods: []string{"POST"},
			},
			Description:    "Create users rule",
			Authorizer:     rule.Handler{Handler: "allow", Config: json.RawMessage(`{"type":"any"}`)},
			Authenticators: []rule.Handler{{Handler: "anonymous", Config: json.RawMessage(`{"name":"anonymous1"}`)}},
			Mutators:       []rule.Handler{{Handler: "id_token", Config: json.RawMessage(`{"issuer":"anything"}`)}},
			Upstream: rule.Upstream{
				URL:          "http://localhost:1235/",
				StripPath:    "/bar",
				PreserveHost: true,
			},
		},
		{
			ID: "foo2",
			Match: &rule.Match{
				URL:     "https://localhost:34/<{baz*,bar*}>",
				Methods: []string{"GET"},
			},
			Description:    "Get users rule",
			Authorizer:     rule.Handler{Handler: "deny", Config: json.RawMessage(`{"type":"any"}`)},
			Authenticators: []rule.Handler{{Handler: "oauth2_introspection", Config: json.RawMessage(`{"name":"anonymous1"}`)}},
			Mutators:       []rule.Handler{{Handler: "id_token", Config: json.RawMessage(`{"issuer":"anything"}`)}, {Handler: "headers", Config: json.RawMessage(`{"headers":{"X-User":"user"}}`)}},
			Upstream: rule.Upstream{
				URL:          "http://localhost:333/",
				StripPath:    "/foo",
				PreserveHost: false,
			},
		},
	}

	t.Run("case=create a new rule", func(t *testing.T) {
		testFunc := func(strategy configuration.MatchingStrategy, rules []rule.Rule) {
			reg.RuleRepository().(*rule.RepositoryMemory).WithRules(rules)
			require.NoError(t, reg.RuleRepository().SetMatchingStrategy(context.Background(), strategy))
			results, err := cl.API.ListRules(sdkrule.NewListRulesParams().WithLimit(pointerx.Int64(10)))
			require.NoError(t, err)
			require.Len(t, results.Payload, 2)
			assert.True(t, results.Payload[0].ID != results.Payload[1].ID)

			result, err := cl.API.GetRule(sdkrule.NewGetRuleParams().WithID(rules[1].ID))
			require.NoError(t, err)

			var b bytes.Buffer
			var ruleResult rule.Rule
			require.NoError(t, json.NewEncoder(&b).Encode(result.Payload))
			require.NoError(t, json.NewDecoder(&b).Decode(&ruleResult))

			assert.EqualValues(t, rules[1], ruleResult)
		}
		t.Run("regexp", func(t *testing.T) {
			testFunc(configuration.Regexp, rulesRegexp)
		})
		t.Run("glob", func(t *testing.T) {
			testFunc(configuration.Glob, rulesGlob)
		})

	})
}
