package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/x"
)

func TestHealth(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	r := internal.NewRegistry(conf)

	router := x.NewAPIRouter()
	r.HealthHandler().SetRoutes(router.Router, true)
	server := httptest.NewServer(router)
	defer server.Close()

	var result struct {
		// Status always contains "ok".
		Status string `json:"status"`
	}

	res, err := server.Client().Get(server.URL + "/health/alive")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
	assert.Equal(t, "ok", result.Status)

	res, err = server.Client().Get(server.URL + "/health/ready")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
	assert.Equal(t, "ok", result.Status)
}
