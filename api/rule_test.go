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

package api

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ory/oathkeeper/sdk/go/oathkeeper/client"
	"github.com/ory/oathkeeper/sdk/go/oathkeeper/client/rule"
	"github.com/ory/oathkeeper/sdk/go/oathkeeper/models"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/pkg"
)

func TestHandler(t *testing.T) {
	handler := &Handler{
		H: herodot.NewJSONWriter(nil),
		M: NewMemoryManager(),
		V: ValidateRule(
			[]string{"anonymous", "oauth2_introspection"}, []string{"anonymous", "oauth2_introspection"},
			[]string{"allow", "deny"}, []string{"allow", "deny"},
			[]string{"id_token"}, []string{"id_token"},
		),
	}
	router := httprouter.New()
	handler.SetRoutes(router)
	server := httptest.NewServer(router)

	u, err := url.ParseRequestURI(server.URL)
	require.NoError(t, err)

	cl := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:     u.Host,
		BasePath: u.Path,
		Schemes:  []string{u.Scheme},
	})

	r1 := models.SwaggerRule{
		ID: "foo1",
		Match: &models.SwaggerRuleMatch{
			URL:     "https://localhost:1234/<foo|bar>",
			Methods: []string{"POST"},
		},
		Description:       "Create users rule",
		Authorizer:        &models.SwaggerRuleHandler{Handler: "allow", Config: map[string]interface{}{"type": "any"}},
		Authenticators:    []*models.SwaggerRuleHandler{{Handler: "anonymous", Config: map[string]interface{}{"name": "anonymous1"}}},
		CredentialsIssuer: &models.SwaggerRuleHandler{Handler: "id_token", Config: map[string]interface{}{"issuer": "anything"}},
		Upstream: &models.Upstream{
			URL:          "http://localhost:1235/",
			StripPath:    "/bar",
			PreserveHost: true,
		},
	}
	r2 := models.SwaggerRule{
		ID: "foo2",
		Match: &models.SwaggerRuleMatch{
			URL:     "https://localhost:34/<baz|bar>",
			Methods: []string{"GET"},
		},
		Description:       "Get users rule",
		Authorizer:        &models.SwaggerRuleHandler{Handler: "deny", Config: map[string]interface{}{"type": "any"}},
		Authenticators:    []*models.SwaggerRuleHandler{{Handler: "oauth2_introspection", Config: map[string]interface{}{"name": "anonymous1"}}},
		CredentialsIssuer: &models.SwaggerRuleHandler{Handler: "id_token", Config: map[string]interface{}{"issuer": "anything"}},
		Upstream: &models.Upstream{
			URL:          "http://localhost:333/",
			StripPath:    "/foo",
			PreserveHost: false,
		},
	}
	invalidRule := models.SwaggerRule{
		ID: "foo3",
	}

	t.Run("case=create a new rule", func(t *testing.T) {
		_, err := cl.Rule.CreateRule(rule.NewCreateRuleParams().WithBody(&invalidRule))
		require.Error(t, err)

		result, err := cl.Rule.CreateRule(rule.NewCreateRuleParams().WithBody(&r1))
		require.NoError(t, err)
		assert.EqualValues(t, r1, *result.Payload)

		result, err = cl.Rule.CreateRule(rule.NewCreateRuleParams().WithBody(&r2))
		require.NoError(t, err)
		assert.NotEmpty(t, result.Payload.ID)
		r2.ID = result.Payload.ID

		results, err := cl.Rule.ListRules(rule.NewListRulesParams().WithLimit(&pkg.RulesUpperLimit))
		require.NoError(t, err)
		require.Len(t, results.Payload, 2)
		assert.True(t, results.Payload[0].ID != results.Payload[1].ID)

		r1.ID = "newfoo"
		uresult, err := cl.Rule.UpdateRule(rule.NewUpdateRuleParams().WithID("foo1").WithBody(&r1))
		require.NoError(t, err)
		assert.Equal(t, "foo1", uresult.Payload.ID)
		r1.ID = "foo1"

		gresult, err := cl.Rule.GetRule(rule.NewGetRuleParams().WithID(r2.ID))
		require.NoError(t, err)
		assert.EqualValues(t, r2, *gresult.Payload)

		_, err = cl.Rule.DeleteRule(rule.NewDeleteRuleParams().WithID(r1.ID))
		require.NoError(t, err)

		_, err = cl.Rule.GetRule(rule.NewGetRuleParams().WithID(r1.ID))
		require.Error(t, err)

		results, err = cl.Rule.ListRules(rule.NewListRulesParams().WithLimit(&pkg.RulesUpperLimit))
		require.NoError(t, err)
		assert.Len(t, results.Payload, 1)
	})
}
