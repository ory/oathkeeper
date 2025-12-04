// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/tidwall/sjson"

	"github.com/ory/x/configx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/x"

	"github.com/ory/oathkeeper/pipeline/authn"
	. "github.com/ory/oathkeeper/pipeline/authz"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/rule"
)

func TestAuthorizerKetoWarden(t *testing.T) {
	t.Parallel()
	conf := internal.NewConfigurationWithDefaults(configx.SkipValidation())
	reg := internal.NewRegistry(conf)

	rule := &rule.Rule{ID: "TestAuthorizer"}

	a, err := reg.PipelineAuthorizer("keto_engine_acp_ory")
	require.NoError(t, err)
	assert.Equal(t, "keto_engine_acp_ory", a.GetID())

	for k, tc := range []struct {
		setup     func(t *testing.T) *httptest.Server
		r         *http.Request
		session   *authn.AuthenticationSession
		config    json.RawMessage
		expectErr bool
	}{
		{
			r:         &http.Request{},
			expectErr: true,
		},
		{
			config:    []byte(`{ "required_action": "action", "required_resource": "resource" }`),
			r:         &http.Request{URL: &url.URL{}},
			session:   new(authn.AuthenticationSession),
			expectErr: true,
		},
		{
			config: []byte(`{ "required_action": "action", "required_resource": "resource", "flavor": "regex" }`),
			r:      &http.Request{URL: &url.URL{}},
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusForbidden)
				}))
			},
			session:   new(authn.AuthenticationSession),
			expectErr: true,
		},
		{
			config: []byte(`{ "required_action": "action", "required_resource": "resource", "flavor": "exact" }`),
			r:      &http.Request{URL: &url.URL{}},
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Contains(t, r.Header, "Content-Type")
					assert.Contains(t, r.Header["Content-Type"], "application/json")
					assert.Contains(t, r.URL.Path, "exact")
					w.Write([]byte(`{"allowed":false}`)) //nolint:errcheck,gosec // test handler ignores errors
				}))
			},
			session:   new(authn.AuthenticationSession),
			expectErr: true,
		},
		{
			config: []byte(`{ "required_action": "action:{{ printIndex .MatchContext.RegexpCaptureGroups (sub 1 1 | int)}}:{{ index .MatchContext.RegexpCaptureGroups (sub 2 1 | int)}}", "required_resource": "resource:{{ index .MatchContext.RegexpCaptureGroups 0}}:{{ index .MatchContext.RegexpCaptureGroups 1}}" }`),
			r:      &http.Request{URL: x.ParseURLOrPanic("https://localhost/api/users/1234/abcde")},
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var ki AuthorizerKetoEngineACPORYRequestBody
					require.NoError(t, json.NewDecoder(r.Body).Decode(&ki))
					assert.EqualValues(t, AuthorizerKetoEngineACPORYRequestBody{
						Action:   "action:1234:abcde",
						Resource: "resource:1234:abcde",
						Context:  map[string]interface{}{},
						Subject:  "peter",
					}, ki)
					assert.Contains(t, r.URL.Path, "regex")
					w.Write([]byte(`{"allowed":true}`)) //nolint:errcheck,gosec // test handler ignores errors
				}))
			},
			session: &authn.AuthenticationSession{
				Subject: "peter",
				MatchContext: authn.MatchContext{
					RegexpCaptureGroups: []string{"1234", "abcde"},
				},
			},
			expectErr: false,
		},
		{
			config: []byte(`{ "required_action": "action:{{ index .MatchContext.RegexpCaptureGroups 0}}:{{ index .MatchContext.RegexpCaptureGroups 1}}", "required_resource": "resource:{{ index .MatchContext.RegexpCaptureGroups 0}}:{{ index .MatchContext.RegexpCaptureGroups 1}}", "subject": "{{ .Extra.name }}" }`),
			r:      &http.Request{URL: x.ParseURLOrPanic("https://localhost/api/users/1234/abcde")},
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var ki AuthorizerKetoEngineACPORYRequestBody
					require.NoError(t, json.NewDecoder(r.Body).Decode(&ki))
					assert.EqualValues(t, AuthorizerKetoEngineACPORYRequestBody{
						Action:   "action:1234:abcde",
						Resource: "resource:1234:abcde",
						Context:  map[string]interface{}{},
						Subject:  "peter",
					}, ki)
					assert.Contains(t, r.URL.Path, "regex")
					w.Write([]byte(`{"allowed":true}`)) //nolint:errcheck,gosec // test handler ignores errors
				}))
			},
			session: &authn.AuthenticationSession{
				Extra: map[string]interface{}{"name": "peter"},
				MatchContext: authn.MatchContext{
					RegexpCaptureGroups: []string{"1234", "abcde"},
				}},
			expectErr: false,
		},
		{
			config: []byte(`{ "required_action": "action:{{ index .MatchContext.RegexpCaptureGroups 0 }}:{{ .Extra.name }}", "required_resource": "resource:{{ index .MatchContext.RegexpCaptureGroups 0}}:{{ .Extra.apiVersion }}", "subject": "{{ .Extra.name }}" }`),
			r:      &http.Request{URL: x.ParseURLOrPanic("https://localhost/api/users/1234/abcde?limit=10")},
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					var ki AuthorizerKetoEngineACPORYRequestBody
					require.NoError(t, json.NewDecoder(r.Body).Decode(&ki))
					assert.EqualValues(t, AuthorizerKetoEngineACPORYRequestBody{
						Action:   "action:1234:peter",
						Resource: "resource:1234:1.0",
						Context:  map[string]interface{}{},
						Subject:  "peter",
					}, ki)
					assert.Contains(t, r.URL.Path, "regex")
					w.Write([]byte(`{"allowed":true}`)) //nolint:errcheck,gosec // test handler ignores errors
				}))
			},
			session: &authn.AuthenticationSession{
				Extra: map[string]interface{}{
					"name":       "peter",
					"apiVersion": "1.0"},
				MatchContext: authn.MatchContext{RegexpCaptureGroups: []string{"1234"}},
			},
			expectErr: false,
		},
	} {
		k := k
		tc := tc
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			t.Parallel()

			baseURL := "http://73fa403f-7e9c-48ef-870f-d21b2c34fc80c6cb6404-bb36-4e70-8b90-45155657fda6/"
			if tc.setup != nil {
				ts := tc.setup(t)
				defer ts.Close()
				baseURL = ts.URL
			}

			a.(*AuthorizerKetoEngineACPORY).WithContextCreator(func(r *http.Request) map[string]interface{} {
				return map[string]interface{}{}
			})

			tc.config, _ = sjson.SetBytes(tc.config, "base_url", baseURL)
			err := a.Authorize(tc.r, tc.session, tc.config, rule)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("method=validate", func(t *testing.T) {
		conf.SetForTest(t, configuration.AuthorizerKetoEngineACPORYIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"base_url":"","required_action":"foo","required_resource":"bar"}`)))

		conf.SetForTest(t, configuration.AuthorizerKetoEngineACPORYIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"base_url":"http://foo/bar","required_action":"foo","required_resource":"bar"}`)))

		conf.SetForTest(t, configuration.AuthorizerKetoEngineACPORYIsEnabled, true)
		require.Error(t, a.Validate(json.RawMessage(`{"base_url":"","required_action":"foo","required_resource":"bar"}`)))

		conf.SetForTest(t, configuration.AuthorizerKetoEngineACPORYIsEnabled, true)
		require.NoError(t, a.Validate(json.RawMessage(`{"base_url":"http://foo/bar","required_action":"foo","required_resource":"bar"}`)))
	})
}
