package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/internal"
	rulereadiness "github.com/ory/oathkeeper/rule/readiness"
	"github.com/ory/oathkeeper/x"
)

type statusResult struct {
	// Status should contains "ok" in case of success
	Status string `json:"status"`
	// Otherwise a map of error messages is returned
	Error *statusError `json:"error"`
}

type statusError struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func TestHealth(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	r := internal.NewRegistry(conf)

	router := x.NewAPIRouter()
	r.HealthHandler().SetHealthRoutes(router.Router, true)
	server := httptest.NewServer(router)
	defer server.Close()

	var result statusResult

	// Checking health state before initializing the registry
	res, err := server.Client().Get(server.URL + "/health/alive")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
	assert.Equal(t, "ok", result.Status)
	assert.Nil(t, result.Error)

	result = statusResult{}
	res, err = server.Client().Get(server.URL + "/health/ready")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
	assert.Empty(t, result.Status)
	assert.Equal(t, rulereadiness.ErrRuleNotYetLoaded.Error(), result.Error.Message)

	r.Init()
	// Waiting for rule load and health event propagation
	time.Sleep(100 * time.Millisecond)

	// Checking health state after initializing the registry
	result = statusResult{}
	res, err = server.Client().Get(server.URL + "/health/alive")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
	assert.Equal(t, "ok", result.Status)
	assert.Nil(t, result.Error)

	result = statusResult{}
	res, err = server.Client().Get(server.URL + "/health/ready")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
	assert.Equal(t, "ok", result.Status)
	assert.Nil(t, result.Error)
}
