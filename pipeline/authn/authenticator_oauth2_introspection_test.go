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
	"sync"
	"testing"
	"time"

	"github.com/ory/x/assertx"

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

	t.Run("method=authenticate-with-cache", func(t *testing.T) {
		viper.Set("authenticators.oauth2_introspection.config.cache.enabled", true)
		t.Cleanup(func() {
			viper.Set("authenticators.oauth2_introspection.config.cache.enabled", false)
		})

		var didNotUseCache sync.WaitGroup

		setup := func(t *testing.T, config string) []byte {
			router := httprouter.New()
			router.POST("/oauth2/introspect", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
				defer didNotUseCache.Done()
				require.NoError(t, r.ParseForm())
				switch r.Form.Get("token") {
				case "inactive-scope-b":
					require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
						Active: false,
					}))
				case "another-active-scope-a":
					fallthrough
				case "active-scope-a":
					if r.Form.Get("scope") != "" && r.Form.Get("scope") != "scope-a" {
						require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
							Active: false,
						}))
						return
					}
					require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
						Active:   true,
						Scope:    "scope-a",
						Subject:  "subject",
						Audience: []string{"audience"},
						Issuer:   "foo",
						Username: "username",
						Expires:  time.Now().Add(time.Second).Unix(),
						Extra:    map[string]interface{}{"extra": "foo"},
					}))
				case "refresh-token":
					require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
						Active:   true,
						Scope:    "scope-a",
						Subject:  "subject",
						Audience: []string{"audience"},
						Issuer:   "foo",
						Username: "username",
						TokenUse: "refresh_token",
						Extra:    map[string]interface{}{"extra": "foo"},
					}))
				default:
					require.NoError(t, json.NewEncoder(w).Encode(&AuthenticatorOAuth2IntrospectionResult{
						Active: false,
					}))
				}
			})
			ts := httptest.NewServer(router)
			t.Cleanup(ts.Close)

			config, err = sjson.Set(config, "introspection_url", ts.URL+"/oauth2/introspect")
			require.NoError(t, err)
			config, err = sjson.Set(config, "pre_authorization.token_url", ts.URL+"/oauth2/token")
			require.NoError(t, err)

			return []byte(config)
		}

		t.Run("case=with none scope strategy", func(t *testing.T) {
			viper.Set("authenticators.oauth2_introspection.config.scope_strategy", "none")
			r := &http.Request{Header: http.Header{"Authorization": {"bearer active-scope-a"}}}
			expected := new(AuthenticationSession)
			t.Run("case=initial request succeeds and caches", func(t *testing.T) {
				config := setup(t, `{ "required_scope": [], "trusted_issuers": ["foo", "bar"], "target_audience": ["audience"] }`)

				didNotUseCache.Add(1)
				err = a.Authenticate(r, expected, config, nil)
				didNotUseCache.Wait()
				require.NoError(t, err)
			})

			// We expect to use the cache here because we are not interested to validate the scope. Usually we would
			// expect to make the upstream call if the upstream has to validate the scope.
			t.Run("case=second request does use cache because no scope was requested and strategy is nil", func(t *testing.T) {
				config := setup(t, `{ "trusted_issuers": ["foo", "bar"], "target_audience": ["audience"] }`)
				sess := new(AuthenticationSession)

				err = a.Authenticate(r, sess, config, nil)
				didNotUseCache.Wait() // Would result in a panic if wg.done was called!
				require.NoError(t, err)
				assertx.EqualAsJSON(t, expected, sess)
			})

			t.Run("case=second request does not use cache because scope strategy is disabled and scope was requested request succeeds", func(t *testing.T) {
				config := setup(t, `{ "required_scope": ["scope-a"], "trusted_issuers": ["foo", "bar"], "target_audience": ["audience"] }`)
				sess := new(AuthenticationSession)

				didNotUseCache.Add(1)
				err = a.Authenticate(r, sess, config, nil)
				didNotUseCache.Wait()
				require.NoError(t, err)
				assertx.EqualAsJSON(t, expected, sess)
			})

			t.Run("case=request fails because we requested a scope which the upstream does not validate", func(t *testing.T) {
				config := setup(t, `{ "required_scope": ["scope-b"], "trusted_issuers": ["foo", "bar"], "target_audience": ["audience"] }`)
				sess := new(AuthenticationSession)

				didNotUseCache.Add(1)
				err = a.Authenticate(r, sess, config, nil)
				didNotUseCache.Wait()
				require.Error(t, err)
			})
		})

		t.Run("case=does not use cache for refresh tokens", func(t *testing.T) {
			for _, strategy := range []string{"wildcard", "none"} {
				t.Run("scope_strategy="+strategy, func(t *testing.T) {
					viper.Set("authenticators.oauth2_introspection.config.scope_strategy", strategy)
					r := &http.Request{Header: http.Header{"Authorization": {"bearer refresh_token"}}}
					expected := new(AuthenticationSession)

					// The initial request
					config := setup(t, `{ "required_scope": ["scope-a"], "trusted_issuers": ["foo", "bar"], "target_audience": ["audience"] }`)

					// Also doesn't use the cache the second time
					didNotUseCache.Add(2)
					require.Error(t, a.Authenticate(r, expected, config, nil))
					require.Error(t, a.Authenticate(r, expected, config, nil))
					didNotUseCache.Wait()
				})
			}
		})

		t.Run("case=with a scope scope strategy", func(t *testing.T) {
			viper.Set("authenticators.oauth2_introspection.config.scope_strategy", "wildcard")
			r := &http.Request{Header: http.Header{"Authorization": {"bearer another-active-scope-a"}}}
			expected := new(AuthenticationSession)

			// The initial request
			config := setup(t, `{ "required_scope": ["scope-a"], "trusted_issuers": ["foo", "bar"], "target_audience": ["audience"] }`)

			didNotUseCache.Add(1)
			require.NoError(t, a.Authenticate(r, expected, config, nil))
			didNotUseCache.Wait()

			t.Run("case=request succeeds and uses the cache", func(t *testing.T) {
				config := setup(t, `{ "trusted_issuers": ["foo", "bar"], "target_audience": ["audience"] }`)
				sess := new(AuthenticationSession)

				err = a.Authenticate(r, sess, config, nil)
				didNotUseCache.Wait()
				require.NoError(t, err)
				assertx.EqualAsJSON(t, expected, sess)
			})

			t.Run("case=request the initial request which also passes", func(t *testing.T) {
				config := setup(t, `{ "required_scope": ["scope-a"], "trusted_issuers": ["foo", "bar"], "target_audience": ["audience"] }`)
				sess := new(AuthenticationSession)

				err = a.Authenticate(r, sess, config, nil)
				didNotUseCache.Wait()
				require.NoError(t, err)
				assertx.EqualAsJSON(t, expected, sess)
			})

			t.Run("case=requests a scope the token does not have", func(t *testing.T) {
				require.Error(t, a.Authenticate(r, new(AuthenticationSession),
					setup(t, `{ "required_scope": ["scope-b"], "trusted_issuers": ["foo", "bar"], "target_audience": ["audience"] }`),
					nil))
				didNotUseCache.Wait()
			})

			t.Run("case=requests an audience which the token does not have", func(t *testing.T) {
				require.Error(t, a.Authenticate(r, new(AuthenticationSession),
					setup(t, `{ "required_scope": ["scope-a"], "trusted_issuers": ["foo", "bar"], "target_audience": ["not-audience"] }`),
					nil))
				didNotUseCache.Wait()
			})

			t.Run("case=does not trust the issuer", func(t *testing.T) {
				require.Error(t, a.Authenticate(r, new(AuthenticationSession),
					setup(t, `{ "required_scope": ["scope-a"], "trusted_issuers": ["not-foo", "bar"], "target_audience": ["audience"] }`),
					nil))
				didNotUseCache.Wait()
			})

			t.Run("case=respects the expiry time", func(t *testing.T) {
				setup(t, `{ "required_scope": ["scope-a"], "trusted_issuers": ["foo", "bar"], "target_audience": ["audience"] }`)
				require.NoError(t, a.Authenticate(r, new(AuthenticationSession), config, nil))
				time.Sleep(time.Second)
				require.Error(t, a.Authenticate(r, new(AuthenticationSession), config, nil))
				didNotUseCache.Wait()
			})
		})
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
		require.NoError(t, err)

		_, preauthOneClient, err := authenticator.Config(preAuthConfigOne)
		require.NoError(t, err)

		_, preauthTwoClient, err := authenticator.Config(preAuthConfigTwo)
		require.NoError(t, err)

		require.NotEqual(t, noPreauthClient, preauthOneClient)
		require.NotEqual(t, noPreauthClient, preauthTwoClient)
		require.NotEqual(t, preauthOneClient, preauthTwoClient)

		_, noPreauthClient2, err := authenticator.Config(noPreauthConfig)
		require.NoError(t, err)
		require.Equal(t, noPreauthClient2, noPreauthClient)

		_, preauthOneClient2, err := authenticator.Config(preAuthConfigOne)
		require.NoError(t, err)
		require.Equal(t, preauthOneClient2, preauthOneClient)

		_, preauthTwoClient2, err := authenticator.Config(preAuthConfigTwo)
		require.NoError(t, err)
		require.Equal(t, preauthTwoClient2, preauthTwoClient)

		require.NotEqual(t, noPreauthClient2, preauthOneClient)
		require.NotEqual(t, noPreauthClient2, preauthTwoClient)

		require.NotEqual(t, preauthOneClient2, noPreauthClient)
		require.NotEqual(t, preauthOneClient2, preauthTwoClient)

		require.NotEqual(t, preauthTwoClient2, noPreauthClient)
		require.NotEqual(t, preauthTwoClient2, preauthOneClient)

		t.Run("Should not be equal because we changed a system default", func(t *testing.T) {
			// Unskip once https://github.com/ory/oathkeeper/issues/757 lands
			t.Skip("This fails due to viper caching and it makes no sense to fix it as we need to adopt koanf first")
			viper.Set("authenticators.oauth2_introspection.config.pre_authorization", map[string]interface{}{"scope": []string{"foo"}})

			_, noPreauthClient3, err := authenticator.Config(noPreauthConfig)
			require.NoError(t, err)
			require.NotEqual(t, noPreauthClient3, noPreauthClient)
		})
	})
}
