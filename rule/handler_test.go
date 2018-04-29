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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/pkg"
	"github.com/ory/oathkeeper/sdk/go/oathkeeper/swagger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	handler := &Handler{
		H: herodot.NewJSONWriter(nil),
		M: NewMemoryManager(),
	}
	router := httprouter.New()
	handler.SetRoutes(router)
	server := httptest.NewServer(router)

	client := swagger.NewRuleApiWithBasePath(server.URL)

	r1 := swagger.Rule{
		Id: "foo1",
		Match: swagger.RuleMatch{
			Url:     "https://localhost:1234/<foo|bar>",
			Methods: []string{"POST"},
		},
		Description:       "Create users rule",
		Authorizer:        swagger.RuleHandler{Handler: "allow", Config: []byte(`{"type":"any"}`)},
		Authenticators:    []swagger.RuleHandler{{Handler: "anonymous", Config: []byte(`{"name":"anonymous1"}`)}},
		CredentialsIssuer: swagger.RuleHandler{Handler: "id_token", Config: []byte(`{"issuer":"anything"}`)},
		Upstream: swagger.Upstream{
			Url:          "http://localhost:1235/",
			StripPath:    "/bar",
			PreserveHost: true,
		},
	}
	r2 := swagger.Rule{
		Id: "foo2",
		Match: swagger.RuleMatch{
			Url:     "https://localhost:34/<baz|bar>",
			Methods: []string{"GET"},
		},
		Description:       "Get users rule",
		Authorizer:        swagger.RuleHandler{Handler: "deny", Config: []byte(`{"type":"any"}`)},
		Authenticators:    [] swagger.RuleHandler{{Handler: "oauth2_introspection", Config: []byte(`{"name":"anonymous1"}`)}},
		CredentialsIssuer: swagger.RuleHandler{Handler: "id_token", Config: []byte(`{"issuer":"anything"}`)},
		Upstream: swagger.Upstream{
			Url:          "http://localhost:333/",
			StripPath:    "/foo",
			PreserveHost: false,
		},
	}

	t.Run("case=create a new rule", func(t *testing.T) {
		result, response, err := client.CreateRule(r1)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.StatusCode)
		assert.EqualValues(t, r1, *result)

		result, response, err = client.CreateRule(r2)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.StatusCode)
		assert.NotEmpty(t, result.Id)
		r2.Id = result.Id

		results, response, err := client.ListRules(pkg.RulesUpperLimit, 0)
		require.NoError(t, err)
		require.Len(t, results, 2)
		assert.True(t, results[0].Id != results[1].Id)

		r1.Id = "newfoo"
		result, response, err = client.UpdateRule("foo1", r1)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "foo1", result.Id)
		r1.Id = "foo1"

		result, response, err = client.GetRule(r2.Id)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.EqualValues(t, r2, *result)

		response, err = client.DeleteRule(r1.Id)
		require.NoError(t, err)

		_, response, err = client.GetRule(r1.Id)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, response.StatusCode)

		results, response, err = client.ListRules(pkg.RulesUpperLimit, 0)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}
