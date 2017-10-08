package rule

import (
	"net/http/httptest"
	"testing"

	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/sdk/swagger"
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
		Id:               "foo1",
		Description:      "Create users rule",
		MatchesPath:      "/users/([0-9]+)",
		MatchesMethods:   []string{"POST"},
		RequiredResource: "users:$1",
		RequiredAction:   "create:$1",
		RequiredScopes:   []string{"users.create"},
	}
	r2 := swagger.Rule{
		Description:         "Get users rule",
		MatchesPath:         "/users/([0-9]+)",
		MatchesMethods:      []string{"GET"},
		AllowAnonymous:      true,
		BypassAuthorization: true,
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

		results, response, err := client.ListRules()
		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.True(t, results[0].Id != results[1].Id)

		r1.RequiredScopes = []string{"users"}
		r1.Id = "newfoo"
		result, response, err = client.UpdateRule("foo1", r1)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "foo1", result.Id)
		assert.EqualValues(t, r1.RequiredScopes, result.RequiredScopes)
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

		results, response, err = client.ListRules()
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}
