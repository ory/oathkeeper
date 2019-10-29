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

package authz_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/tidwall/sjson"

	"github.com/ory/viper"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"

	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
	. "github.com/ory/oathkeeper/pipeline/authz"

	"github.com/ory/x/urlx"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/rule"
)

func TestAuthorizerKetoWarden(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineAuthorizer("keto_engine_acp_ory")
	require.NoError(t, err)
	assert.Equal(t, "keto_engine_acp_ory", a.GetID())

	for k, tc := range []struct {
		setup     func(t *testing.T) *httptest.Server
		r         *http.Request
		session   *authn.AuthenticationSession
		config    json.RawMessage
		rule      pipeline.Rule
		expectErr bool
	}{
		{
			expectErr: true,
		},
		{
			config: []byte(`{ "required_action": "action", "required_resource": "resource" }`),
			rule: &rule.Rule{
				Match: rule.RuleMatch{
					Methods: []string{"POST"},
					URL:     "https://localhost/",
				},
			},
			r:         &http.Request{URL: &url.URL{}},
			session:   new(authn.AuthenticationSession),
			expectErr: true,
		},
		{
			config: []byte(`{ "required_action": "action", "required_resource": "resource", "flavor": "regex" }`),
			rule: &rule.Rule{
				Match: rule.RuleMatch{
					Methods: []string{"POST"},
					URL:     "https://localhost/",
				},
			},
			r: &http.Request{URL: &url.URL{}},
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusForbidden)
				}))
			},
			session:   new(authn.AuthenticationSession),
			expectErr: true,
		},
		{
			config: []byte(`{ "required_action": "action", "required_resource": "resource", "flavor": "exact" }`),
			rule: &rule.Rule{
				Match: rule.RuleMatch{
					Methods: []string{"POST"},
					URL:     "https://localhost/",
				},
			},
			r: &http.Request{URL: &url.URL{}},
			setup: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Contains(t, r.Header, "Content-Type")
					assert.Contains(t, r.Header["Content-Type"], "application/json")
					assert.Contains(t, r.URL.Path, "exact")
					w.Write([]byte(`{"allowed":false}`))
				}))
			},
			session:   new(authn.AuthenticationSession),
			expectErr: true,
		},
		{
			config: []byte(`{ "required_action": "action:$1:$2", "required_resource": "resource:$1:$2" }`),
			rule: &rule.Rule{
				Match: rule.RuleMatch{
					Methods: []string{"POST"},
					URL:     "https://localhost/api/users/<[0-9]+>/<[a-z]+>",
				},
			},
			r: &http.Request{URL: urlx.ParseOrPanic("https://localhost/api/users/1234/abcde")},
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
					w.Write([]byte(`{"allowed":true}`))
				}))
			},
			session:   &authn.AuthenticationSession{Subject: "peter"},
			expectErr: false,
		},
		{
			config: []byte(`{ "required_action": "action:$1:$2", "required_resource": "resource:$1:$2", "subject": "{{ .Extra.name }}" }`),
			rule: &rule.Rule{
				Match: rule.RuleMatch{
					Methods: []string{"POST"},
					URL:     "https://localhost/api/users/<[0-9]+>/<[a-z]+>",
				},
			},
			r: &http.Request{URL: urlx.ParseOrPanic("https://localhost/api/users/1234/abcde")},
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
					w.Write([]byte(`{"allowed":true}`))
				}))
			},
			session:   &authn.AuthenticationSession{Extra: map[string]interface{}{"name": "peter"}},
			expectErr: false,
		},
		{
			config: []byte(`{ "required_action": "action:$1:$2", "required_resource": "resource:$1:$2", "subject": "{{ .Extra.name }}" }`),
			rule: &rule.Rule{
				Match: rule.RuleMatch{
					Methods: []string{"POST"},
					URL:     "https://localhost/api/users/<[0-9]+>/<[a-z]+>",
				},
			},
			r: &http.Request{URL: urlx.ParseOrPanic("https://localhost/api/users/1234/abcde?limit=10")},
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
					w.Write([]byte(`{"allowed":true}`))
				}))
			},
			session:   &authn.AuthenticationSession{Extra: map[string]interface{}{"name": "peter"}},
			expectErr: false,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

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
			err := a.Authorize(tc.r, tc.session, tc.config, tc.rule)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("method=validate", func(t *testing.T) {
		viper.Set(configuration.ViperKeyAuthorizerKetoEngineACPORYIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"base_url":"","required_action":"foo","required_resource":"bar"}`)))

		viper.Set(configuration.ViperKeyAuthorizerKetoEngineACPORYIsEnabled, true)
		require.Error(t, a.Validate(json.RawMessage(`{"base_url":"","required_action":"foo","required_resource":"bar"}`)))

		viper.Set(configuration.ViperKeyAuthorizerKetoEngineACPORYIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"base_url":"http://foo/bar","required_action":"foo","required_resource":"bar"}`)))

		viper.Set(configuration.ViperKeyAuthorizerKetoEngineACPORYIsEnabled, true)
		require.NoError(t, a.Validate(json.RawMessage(`{"base_url":"http://foo/bar","required_action":"foo","required_resource":"bar"}`)))
	})
}
