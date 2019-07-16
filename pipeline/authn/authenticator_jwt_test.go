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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/ory/viper"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	. "github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/x/urlx"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticatorJWT(t *testing.T) {
	keys := []string{
		"file://../../test/stub/jwks-hs.json",
		"file://../../test/stub/jwks-rsa-multiple.json",
		"file://../../test/stub/jwks-rsa-single.json",
		"file://../../test/stub/jwks-ecdsa.json",
	}
	conf := internal.NewConfigurationWithDefaults()
	viper.Set(configuration.ViperKeyAuthenticatorJWTJWKSURIs, keys)
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineAuthenticator("jwt")
	require.NoError(t, err)
	assert.Equal(t, "jwt", a.GetID())

	var gen = func(l string, c jwt.Claims) string {
		token, err := reg.CredentialsSigner().Sign(context.Background(), urlx.ParseOrPanic(l), c)
		require.NoError(t, err)
		return token
	}

	now := time.Now().UTC()

	t.Run("method=authenticate", func(t *testing.T) {
		for k, tc := range []struct {
			setup      func()
			d          string
			r          *http.Request
			config     string
			expectErr  bool
			expectSess *AuthenticationSession
		}{
			{
				d:         "should fail because no payloads",
				r:         &http.Request{Header: http.Header{}},
				expectErr: true,
			},
			{
				d:         "should fail because not a jwt",
				r:         &http.Request{Header: http.Header{"Authorization": []string{"bearer invalid.token.sign"}}},
				expectErr: true,
			},
			{
				d: "should pass because JWT is valid",
				setup: func() {
					viper.Set(configuration.ViperKeyAuthenticatorJWTScopeStrategy, "exact")
				},
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[1], jwt.MapClaims{
					"sub":   "sub",
					"exp":   now.Add(time.Hour).Unix(),
					"aud":   []string{"aud-1", "aud-2"},
					"iss":   "iss-2",
					"scope": []string{"scope-3", "scope-2", "scope-1"},
				})}}},
				config:    `{"target_audience": ["aud-1", "aud-2"], "trusted_issuers": ["iss-1", "iss-2"], "required_scope": ["scope-1", "scope-2"]}`,
				expectErr: false,
				expectSess: &AuthenticationSession{
					Subject: "sub",
					Extra: map[string]interface{}{
						"sub": "sub",
						"exp": float64(now.Add(time.Hour).Unix()),
						"aud": []interface{}{"aud-1", "aud-2"},
						"iss": "iss-2",
						"scp": []string{"scope-3", "scope-2", "scope-1"},
					},
				},
			},
			{
				d: "should pass because JWT scope can be a string",
				setup: func() {
					viper.Set(configuration.ViperKeyAuthenticatorJWTScopeStrategy, "exact")
				},
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[2], jwt.MapClaims{
					"sub":   "sub",
					"exp":   now.Add(time.Hour).Unix(),
					"aud":   []string{"aud-1", "aud-2"},
					"iss":   "iss-2",
					"scope": "scope-3 scope-2 scope-1",
				})}}},
				config:    `{"target_audience": ["aud-1", "aud-2"], "trusted_issuers": ["iss-1", "iss-2"], "required_scope": ["scope-1", "scope-2"]}`,
				expectErr: false,
				expectSess: &AuthenticationSession{
					Subject: "sub",
					Extra: map[string]interface{}{
						"sub": "sub",
						"exp": float64(now.Add(time.Hour).Unix()),
						"aud": []interface{}{"aud-1", "aud-2"},
						"iss": "iss-2",
						"scp": []string{"scope-3", "scope-2", "scope-1"},
					},
				},
			},
			{
				d: "should pass because JWT is valid and HS256 is allowed",
				setup: func() {
					viper.Set(configuration.ViperKeyAuthenticatorJWTScopeStrategy, "none")
				},
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[0], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
				})}}},
				expectErr: false,
				config:    `{ "allowed_algorithms": ["HS256"] }`,
				expectSess: &AuthenticationSession{
					Subject: "sub",
					Extra:   map[string]interface{}{"scp": []string{}, "sub": "sub", "exp": float64(now.Add(time.Hour).Unix())},
				},
			},
			{
				d: "should pass because JWT is valid and ES256 is allowed",
				setup: func() {
					viper.Set(configuration.ViperKeyAuthenticatorJWTScopeStrategy, "none")
				},
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[3], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
				})}}},
				expectErr: false,
				config:    `{ "allowed_algorithms": ["ES256"] }`,
				expectSess: &AuthenticationSession{
					Subject: "sub",
					Extra:   map[string]interface{}{"scp": []string{}, "sub": "sub", "exp": float64(now.Add(time.Hour).Unix())},
				},
			},
			{
				d: "should pass because JWT is valid",
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[1], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
				})}}},
				config:    `{}`,
				expectErr: false,
			},
			{
				d: "should fail because JWT nbf is in future",
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[2], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
					"nbf": now.Add(time.Hour).Unix(),
				})}}},
				config:    `{}`,
				expectErr: true,
			},
			{
				d: "should fail because JWT iat is in future",
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[2], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
					"iat": now.Add(time.Hour).Unix(),
				})}}},
				config:    `{}`,
				expectErr: true,
			},
			{
				d: "should pass because JWT is missing scope",
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[2], jwt.MapClaims{
					"sub":   "sub",
					"exp":   now.Add(time.Hour).Unix(),
					"scope": []string{"scope-1", "scope-2"},
				})}}},
				config:    `{"required_scope": ["scope-1", "scope-2", "scope-3"]}`,
				expectErr: true,
			},
			{
				d: "should pass because JWT issuer is untrusted",
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[1], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
					"iss": "iss-4",
				})}}},
				config:    `{"trusted_issuers": ["iss-1", "iss-2", "iss-3"]}`,
				expectErr: true,
			},
			{
				d: "should pass because JWT is missing audience",
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[1], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
					"aud": []string{"aud-1", "aud-2"},
				})}}},
				config:    `{"required_audience": ["aud-1", "aud-2", "aud-3"]}`,
				expectErr: true,
			},
			{
				d: "should fail because JWT is expired",
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[1], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(-time.Hour).Unix(),
				})}}},
				config:    `{}`,
				expectErr: true,
			},
		} {
			t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
				if tc.setup != nil {
					tc.setup()
				}

				session, err := a.Authenticate(tc.r, json.RawMessage([]byte(tc.config)), nil)
				if tc.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err, "%#v", errors.Cause(err))
				}

				if tc.expectSess != nil {
					assert.Equal(t, tc.expectSess, session)
				}
			})
		}
	})

	t.Run("method=validate", func(t *testing.T) {

	})
}
