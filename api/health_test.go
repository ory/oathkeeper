// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/x/configx"
	"github.com/ory/x/httprouterx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
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
	r := internal.NewRegistry(t, configx.WithValues(map[string]any{
		configuration.AccessRuleRepositories:          []string{"file://../test/stub/rules.json"},
		configuration.AuthorizerAllowIsEnabled:        true,
		configuration.AuthorizerDenyIsEnabled:         true,
		configuration.AuthenticatorNoopIsEnabled:      true,
		configuration.AuthenticatorAnonymousIsEnabled: true,
		configuration.MutatorNoopIsEnabled:            true,
		"mutators.header.config":                      map[string]interface{}{"headers": map[string]interface{}{}},
		configuration.MutatorHeaderIsEnabled:          true,
		configuration.MutatorIDTokenJWKSURL:           "https://stub/.well-known/jwks.json",
		configuration.MutatorIDTokenIssuerURL:         "https://stub",
		configuration.MutatorIDTokenIsEnabled:         true,
	}))

	router := httprouterx.NewRouter()
	r.HealthHandler().SetHealthRoutes(router, true)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	var result statusResult

	// Checking health state before initializing the registry
	res, err := server.Client().Get(server.URL + "/health/alive")
	require.NoError(t, err)
	t.Cleanup(func(b io.Closer) func() { return func() { _ = b.Close() } }(res.Body))
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
	assert.Equal(t, "ok", result.Status)
	assert.Nil(t, result.Error)

	result = statusResult{}
	res, err = server.Client().Get(server.URL + "/health/ready")
	require.NoError(t, err)
	t.Cleanup(func(b io.Closer) func() { return func() { _ = b.Close() } }(res.Body))
	require.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
	assert.Empty(t, result.Status)
	assert.Equal(t, herodot.ErrNotFound().ErrorField, result.Error.Message)

	r.Init()
	// Waiting for rule load and health event propagation
	time.Sleep(100 * time.Millisecond)

	// Checking health state after initializing the registry
	result = statusResult{}
	res, err = server.Client().Get(server.URL + "/health/alive")
	require.NoError(t, err)
	t.Cleanup(func(b io.Closer) func() { return func() { _ = b.Close() } }(res.Body))
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
	assert.Equal(t, "ok", result.Status)
	assert.Nil(t, result.Error)

	result = statusResult{}
	res, err = server.Client().Get(server.URL + "/health/ready")
	require.NoError(t, err)
	t.Cleanup(func(b io.Closer) func() { return func() { _ = b.Close() } }(res.Body))
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
	assert.Equal(t, "ok", result.Status)
	assert.Nil(t, result.Error)
}
