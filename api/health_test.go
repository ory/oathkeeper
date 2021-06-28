// Copyright 2021 Ory GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	Errors map[string]string `json:"errors"`
}

func TestHealth(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	r := internal.NewRegistry(conf)

	router := x.NewAPIRouter()
	r.HealthHandler().SetRoutes(router.Router, true)
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
	assert.Len(t, result.Errors, 0)

	result = statusResult{}
	res, err = server.Client().Get(server.URL + "/health/ready")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
	assert.Empty(t, result.Status)
	assert.Len(t, result.Errors, 1)
	assert.Equal(t, rulereadiness.ErrRuleNotYetLoaded.Error(), result.Errors[rulereadiness.ProbeName])

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
	assert.Len(t, result.Errors, 0)

	result = statusResult{}
	res, err = server.Client().Get(server.URL + "/health/ready")
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NoError(t, json.NewDecoder(res.Body).Decode(&result))
	assert.Equal(t, "ok", result.Status)
	assert.Len(t, result.Errors, 0)
}
