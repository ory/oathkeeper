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

package authn_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	. "github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/viper"
	"github.com/ory/x/logrusx"
)

func TestAuthenticatorOAuth2Introspection(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineAuthenticator("oauth2_introspection")
	require.NoError(t, err)
	assert.Equal(t, "oauth2_introspection", a.GetID())

	t.Run("method=authenticate", func(t *testing.T) {

		for k, tc := range []struct {
			d              string
			setup          func(*testing.T, *httprouter.Router)
			r              *http.Request
			config         json.RawMessage
			expectErr      bool
			expectExactErr error
			expectSess     *AuthenticationSession
		}{
			{
				d:         "should fail because no payloads",
				r:         &http.Request{Header: http.Header{}},
				expectErr: true,
			},
			{
				d:      "should fail because wrong response",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "required_scope": ["scope-a", "scope-b"] }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.Equal(t, "scope-a scope-b", r.Form.Get("scope"))
						w.WriteHeader(http.StatusNotFound)
					})
				},
				expectErr: true,
			},
			{
				d:              "should return error saying that authenticator is not responsible for validating the request, as the token was not provided in a proper location (default)",
				r:              &http.Request{Header: http.Header{"Foobar": {"bearer token"}}},
				expectErr:      true,
				expectExactErr: ErrAuthenticatorNotResponsible,
			},
			{
				d:              "should return error saying that authenticator is not responsible for validating the request, as the token was not provided in a proper location (custom header)",
				r:              &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config:         []byte(`{"token_from": {"header": "X-Custom-Header"}}`),
				expectErr:      true,
				expectExactErr: ErrAuthenticatorNotResponsible,
			},
			{
				d: "should return error saying that authenticator is not responsible for validating the request, as the token was not provided in a proper location (custom query parameter)",
				r: &http.Request{
					Form: map[string][]string{
						"someOtherQueryParam": {"token"},
					},
					Header: http.Header{"Authorization": {"bearer token"}},
				},
				config:         []byte(`{"token_from": {"query_parameter": "token"}}`),
				expectErr:      true,
				expectExactErr: ErrAuthenticatorNotResponsible,
			},
			{
				d:              "should return error saying that authenticator is not responsible for validating the request, as the token was not provided in a proper location (cookie)",
				r:              &http.Request{Header: http.Header{"Cookie": {"biscuit=bearer token"}}},
				config:         []byte(`{"token_from": {"cookie": "cake"}}`),
				expectErr:      true,
				expectExactErr: ErrAuthenticatorNotResponsible,
			},
			{
				d:         "should pass because the valid token was provided in a proper location (custom header)",
				r:         &http.Request{Header: http.Header{"X-Custom-Header": {"token"}}},
				config:    []byte(`{"token_from": {"header": "X-Custom-Header"}}`),
				expectErr: false,
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active: true,
						}))
					})
				},
			},
			{
				d: "should pass because the valid token was provided in a proper location (custom query parameter)",
				r: &http.Request{
					Form: map[string][]string{
						"token": {"token"},
					},
				},
				config:    []byte(`{"token_from": {"query_parameter": "token"}}`),
				expectErr: false,
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active: true,
						}))
					})
				},
			},
			{
				d:         "should pass because the valid token was provided in a proper location (cookie)",
				r:         &http.Request{Header: http.Header{"Cookie": {"biscuit=token"}}},
				config:    []byte(`{"token_from": {"cookie": "biscuit"}}`),
				expectErr: false,
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active: true,
						}))
					})
				},
			},
			{
				d:      "should fail because not active",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{}`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   false,
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "issuer",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
						}))
					})
				},
				expectErr: true,
			},
			{
				d:      "should pass because active and no issuer / audience expected",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{}`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "issuer",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
						}))
					})
				},
				expectErr: false,
			},
			{
				d:      "should pass because no scope strategy",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "required_scope": ["scope-a", "scope-b"] }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.Equal(t, "scope-a scope-b", r.Form.Get("scope"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "issuer",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
							Scope:    "scope-z",
						}))
					})
				},
				expectErr: false,
			},
			{
				d:      "should pass because scope strategy is `none`",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "scope_strategy": "none", "required_scope": ["scope-a", "scope-b"] }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.Equal(t, "scope-a scope-b", r.Form.Get("scope"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "issuer",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
							Scope:    "scope-z",
						}))
					})
				},
				expectErr: false,
			},
			{
				d:      "should pass because active and scope matching exactly",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "scope_strategy": "exact", "required_scope": ["scope-a", "scope-b"] }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.Equal(t, "", r.Form.Get("scope"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "issuer",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
							Scope:    "scope-a scope-b",
						}))
					})
				},
				expectErr: false,
			},
			{
				d:      "should fail because active but scope not matching exactly",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "scope_strategy": "exact", "required_scope": ["scope-a", "scope-b", "scope-c"] }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.Equal(t, "", r.Form.Get("scope"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "issuer",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
							Scope:    "scope-a scope-b",
						}))
					})
				},
				expectErr: true,
			},
			{
				d:      "should pass because active and scope matching hierarchically",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "scope_strategy": "hierarchic", "required_scope": ["scope-a", "scope-b.foo", "scope-c.bar"] }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.Equal(t, "", r.Form.Get("scope"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "issuer",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
							Scope:    "scope-a scope-b scope-c.bar",
						}))
					})
				},
				expectErr: false,
			},
			{
				d:      "should fail because active but scope not matching hierarchically",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "scope_strategy": "hierarchic", "required_scope": ["scope-a", "scope-b.foo", "scope-c.bar"] }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.Equal(t, "", r.Form.Get("scope"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "issuer",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
							Scope:    "scope-a scope-b",
						}))
					})
				},
				expectErr: true,
			},
			{
				d:      "should pass because active and scope matching wildcard",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "scope_strategy": "wildcard", "required_scope": ["scope-a", "scope-b.foo", "scope-c.bar"] }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.Equal(t, "", r.Form.Get("scope"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "issuer",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
							Scope:    "scope-a scope-b.* scope-c.bar",
						}))
					})
				},
				expectErr: false,
			},
			{
				d:      "should fail because active but scope not matching wildcard",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "scope_strategy": "wildcard", "required_scope": ["scope-a", "scope-b.foo", "scope-c.bar"] }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.Equal(t, "", r.Form.Get("scope"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "issuer",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
							Scope:    "scope-a scope-b",
						}))
					})
				},
				expectErr: true,
			},
			{
				d:      "should fail because active but issuer not matching",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "required_scope": ["scope-a"], "trusted_issuers": ["foo", "bar"]}`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Scope:    "scope-a",
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "not-foo",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
						}))
					})
				},
				expectErr: true,
			},
			{
				d:      "should pass because active and issuer matching",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "required_scope": ["scope-a"], "trusted_issuers": ["foo", "bar"]}`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Scope:    "scope-a",
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "foo",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
						}))
					})
				},
				expectErr: false,
			},
			{
				d:      "should fail because active but audience not matching",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "required_scope": ["scope-a"], "trusted_issuers": ["foo", "bar"], "target_audience": ["audience"] }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Scope:    "scope-a",
							Subject:  "subject",
							Audience: []string{"not-audience"},
							Issuer:   "foo",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
						}))
					})
				},
				expectErr: true,
			},
			{
				d:      "should fail because token use not matching",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{}`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "issuer",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
							TokenUse: "any-token-use",
						}))
					})
				},
				expectErr: true,
			},
			{
				d:      "should pass",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "required_scope": ["scope-a"], "trusted_issuers": ["foo", "bar"], "target_audience": ["audience"] }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active:   true,
							Scope:    "scope-a",
							Subject:  "subject",
							Audience: []string{"audience"},
							Issuer:   "foo",
							Username: "username",
							Extra:    map[string]interface{}{"extra": "foo"},
						}))
					})
				},
				expectErr: false,
			},
			{
				d:      "should pass because audience and scopes match configuration",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "pre_authorization":{"client_id":"some_id","client_secret":"some_secret","enabled":true,"scope":["foo","bar"]} }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "", r.Form.Get("audience"))
						require.Equal(t, "foo bar", r.Form.Get("scope"))
						w.Header().Set("Content-type", "application/json; charset=us-ascii")
						require.NoError(t, json.NewEncoder(w).Encode(&map[string]interface{}{"access_token": "foo-token"}))
					})
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.Equal(t, "Bearer foo-token", r.Header.Get("authorization"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active: true,
						}))
					})
				},
				expectErr: false,
			},
			{
				d:      "should pass because audience and scopes match configuration",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "pre_authorization":{"client_id":"some_id","client_secret":"some_secret","enabled":true,"scope":["foo","bar"]} }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "", r.Form.Get("audience"))
						require.Equal(t, "foo bar", r.Form.Get("scope"))
						w.Header().Set("Content-type", "application/json; charset=us-ascii")
						require.NoError(t, json.NewEncoder(w).Encode(&map[string]interface{}{"access_token": "foo-token"}))
					})
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.Equal(t, "Bearer foo-token", r.Header.Get("authorization"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active: true,
						}))
					})
				},
				expectErr: false,
			},
			{
				d:      "should pass because audience and scopes should not be requested",
				r:      &http.Request{Header: http.Header{"Authorization": {"bearer token"}}},
				config: []byte(`{ "pre_authorization":{"client_id":"some_id","client_secret":"some_secret","enabled":true} }`),
				setup: func(t *testing.T, m *httprouter.Router) {
					m.POST("/oauth2/token", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "", r.Form.Get("audience"))
						require.Equal(t, "", r.Form.Get("scope"))
						w.Header().Set("Content-type", "application/json; charset=us-ascii")
						require.NoError(t, json.NewEncoder(w).Encode(&map[string]interface{}{"access_token": "foo-token"}))
					})
					m.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
						require.NoError(t, r.ParseForm())
						require.Equal(t, "token", r.Form.Get("token"))
						require.Equal(t, "Bearer foo-token", r.Header.Get("authorization"))
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active: true,
						}))
					})
				},
				expectErr: false,
			},
		} {
			t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
				router := httprouter.New()
				if tc.setup != nil {
					tc.setup(t, router)
				}
				ts := httptest.NewServer(router)
				defer ts.Close()

				tc.config, _ = sjson.SetBytes(tc.config, "introspection_url", ts.URL+"/oauth2/introspect")
				tc.config, _ = sjson.SetBytes(tc.config, "pre_authorization.token_url", ts.URL+"/oauth2/token")

				sess := new(AuthenticationSession)
				err = a.Authenticate(tc.r, sess, tc.config, nil)
				if tc.expectErr {
					require.Error(t, err)
					if tc.expectExactErr != nil {
						assert.EqualError(t, err, tc.expectExactErr.Error(), "%+v", err)
					}
				} else {
					require.NoError(t, err)
				}

				if tc.expectSess != nil {
					assert.Equal(t, tc.expectSess, sess)
				}
			})
		}
	})

	t.Run("method=validate", func(t *testing.T) {
		viper.Set(configuration.ViperKeyAuthenticatorOAuth2TokenIntrospectionIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"introspection_url":""}`)))

		viper.Reset()
		viper.Set(configuration.ViperKeyAuthenticatorOAuth2TokenIntrospectionIsEnabled, true)
		require.Error(t, a.Validate(json.RawMessage(`{"introspection_url":""}`)))

		viper.Reset()
		viper.Set(configuration.ViperKeyAuthenticatorOAuth2TokenIntrospectionIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"introspection_url":"/oauth2/token"}`)))

		viper.Reset()
		viper.Set(configuration.ViperKeyAuthenticatorOAuth2TokenIntrospectionIsEnabled, true)
		require.Error(t, a.Validate(json.RawMessage(`{"introspection_url":"/oauth2/token"}`)))
	})

	t.Run("method=config", func(t *testing.T) {
		logger := logrusx.New("test", "1")
		authenticator := NewAuthenticatorOAuth2Introspection(conf, logger)

		noPreauthConfig := []byte(`{ "introspection_url":"http://localhost/oauth2/token" }`)
		preAuthConfigOne := []byte(`{ "introspection_url":"http://localhost/oauth2/token","pre_authorization":{"token_url":"http://localhost/oauth2/token","client_id":"some_id","client_secret":"some_secret","enabled":true} }`)
		preAuthConfigTwo := []byte(`{ "introspection_url":"http://localhost/oauth2/token2","pre_authorization":{"token_url":"http://localhost/oauth2/token2","client_id":"some_id2","client_secret":"some_secret2","enabled":true} }`)

		_, noPreauthClient, err := authenticator.Config(noPreauthConfig)
		if err != nil {
			require.NoError(t, err)
		}

		_, preauthOneClient, err := authenticator.Config(preAuthConfigOne)
		if err != nil {
			require.NoError(t, err)
		}

		_, preauthTwoClient, err := authenticator.Config(preAuthConfigTwo)
		if err != nil {
			require.NoError(t, err)
		}

		require.NotEqual(t, noPreauthClient, preauthOneClient)
		require.NotEqual(t, noPreauthClient, preauthTwoClient)
		require.NotEqual(t, preauthOneClient, preauthTwoClient)

		_, preauthOneClient2, err := authenticator.Config(preAuthConfigOne)
		if err != nil {
			require.NoError(t, err)
		}
		if preauthOneClient2 != preauthOneClient {
			t.FailNow()
		}

		_, preauthTwoClient2, err := authenticator.Config(preAuthConfigTwo)
		if err != nil {
			require.NoError(t, err)
		}
		if preauthTwoClient2 != preauthTwoClient {
			t.FailNow()
		}

		_, noPreauthClient2, err := authenticator.Config(noPreauthConfig)
		if err != nil {
			require.NoError(t, err)
		}
		if noPreauthClient2 != noPreauthClient {
			t.FailNow()
		}

		require.NotEqual(t, noPreauthClient, preauthOneClient)
		require.NotEqual(t, noPreauthClient, preauthTwoClient)
		require.NotEqual(t, preauthOneClient, preauthTwoClient)

	})
}
