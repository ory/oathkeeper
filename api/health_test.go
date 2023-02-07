// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/driver/configuration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/internal"
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
	conf.SetForTest(t, configuration.AccessRuleRepositories, []string{"file://../test/stub/rules.json"})
	conf.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
	conf.SetForTest(t, configuration.AuthorizerDenyIsEnabled, true)
	conf.SetForTest(t, configuration.AuthenticatorNoopIsEnabled, true)
	conf.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, true)
	conf.SetForTest(t, configuration.MutatorNoopIsEnabled, true)
	conf.SetForTest(t, "mutators.header.config", map[string]interface{}{"headers": map[string]interface{}{}})
	conf.SetForTest(t, configuration.MutatorHeaderIsEnabled, true)
	conf.SetForTest(t, configuration.MutatorIDTokenJWKSURL, "https://stub/.well-known/jwks.json")
	conf.SetForTest(t, configuration.MutatorIDTokenIssuerURL, "https://stub")
	conf.SetForTest(t, configuration.MutatorIDTokenIsEnabled, true)
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
	assert.Equal(t, herodot.ErrNotFound.ErrorField, result.Error.Message)

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
