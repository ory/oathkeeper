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

package authn

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"

	"github.com/ory/fosite"
)

var keys = map[string]interface{}{"HS256": []byte("some-secret")}

func generateKeys(t *testing.T) {
	rs, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)
	keys["RS256"] = rs
	keys["RS256:public"] = rs.Public()

	es, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	keys["ES256"] = es
	keys["ES256:public"] = es.Public()
}

func generateJWT(t *testing.T, claims jwt.Claims, method string) string {
	var sm jwt.SigningMethod
	var key interface{}
	var kid string

	switch method {
	case "RS256":
		key = keys[method]
		sm = jwt.SigningMethodRS256
		kid = method + ":public"
	case "ES256":
		key = keys[method]
		sm = jwt.SigningMethodES256
		kid = method + ":public"
	case "HS256":
		key = keys[method]
		sm = jwt.SigningMethodHS256
		kid = method
	}

	token := jwt.NewWithClaims(sm, claims)
	token.Header["kid"] = kid
	sign, err := token.SigningString()
	require.NoError(t, err)
	j, err := token.Method.Sign(sign, key)
	require.NoError(t, err)
	return fmt.Sprintf("%s.%s", sign, j)
}

func TestNewAuthenticatorJWT(t *testing.T) {
	_, err := NewAuthenticatorJWT("", fosite.ExactScopeStrategy)
	require.Error(t, err)
	_, err = NewAuthenticatorJWT("foo", fosite.ExactScopeStrategy)
	require.Error(t, err)
}

func TestAuthenticatorJWT(t *testing.T) {
	generateKeys(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{
				{KeyID: "RS256:public", Use: "sig", Key: keys["RS256:public"]},
				{KeyID: "ES256:public", Use: "sig", Key: keys["ES256:public"]},
				{KeyID: "HS256", Use: "sig", Key: keys["HS256"]},
			},
		}))
	}))
	defer ts.Close()

	authenticator, err := NewAuthenticatorJWT(ts.URL, fosite.ExactScopeStrategy)
	require.NoError(t, err)
	assert.NotEmpty(t, authenticator.GetID())
	now := time.Now().Round(time.Second)

	for k, tc := range []struct {
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
			r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + generateJWT(t, jwt.MapClaims{
				"sub":   "sub",
				"exp":   now.Add(time.Hour).Unix(),
				"aud":   []string{"aud-1", "aud-2"},
				"iss":   "iss-2",
				"scope": []string{"scope-3", "scope-2", "scope-1"},
			}, "RS256")}}},
			config:    `{"target_audience": ["aud-1", "aud-2"], "trusted_issuers": ["iss-1", "iss-2"], "required_scope": ["scope-1", "scope-2"]}`,
			expectErr: false,
			expectSess: &AuthenticationSession{
				Subject: "sub",
				Extra: map[string]interface{}{
					"sub":   "sub",
					"exp":   float64(now.Add(time.Hour).Unix()),
					"aud":   []interface{}{"aud-1", "aud-2"},
					"iss":   "iss-2",
					"scope": []interface{}{"scope-3", "scope-2", "scope-1"},
				},
			},
		},
		{
			d: "should pass because JWT scope can be a string",
			r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + generateJWT(t, jwt.MapClaims{
				"sub":   "sub",
				"exp":   now.Add(time.Hour).Unix(),
				"aud":   []string{"aud-1", "aud-2"},
				"iss":   "iss-2",
				"scope": "scope-3 scope-2 scope-1",
			}, "RS256")}}},
			config:    `{"target_audience": ["aud-1", "aud-2"], "trusted_issuers": ["iss-1", "iss-2"], "required_scope": ["scope-1", "scope-2"]}`,
			expectErr: false,
			expectSess: &AuthenticationSession{
				Subject: "sub",
				Extra: map[string]interface{}{
					"sub":   "sub",
					"exp":   float64(now.Add(time.Hour).Unix()),
					"aud":   []interface{}{"aud-1", "aud-2"},
					"iss":   "iss-2",
					"scope": []interface{}{"scope-3", "scope-2", "scope-1"},
				},
			},
		},
		{
			d: "should pass because JWT is valid and HS256 is allowed",
			r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + generateJWT(t, jwt.MapClaims{
				"sub": "sub",
				"exp": now.Add(time.Hour).Unix(),
			}, "HS256")}}},
			expectErr: false,
			config:    `{ "allowed_algorithms": ["HS256"] }`,
			expectSess: &AuthenticationSession{
				Subject: "sub",
				Extra:   map[string]interface{}{"sub": "sub", "exp": float64(now.Add(time.Hour).Unix())},
			},
		},
		{
			d: "should pass because JWT is valid and ES256 is allowed",
			r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + generateJWT(t, jwt.MapClaims{
				"sub": "sub",
				"exp": now.Add(time.Hour).Unix(),
			}, "ES256")}}},
			expectErr: false,
			config:    `{ "allowed_algorithms": ["ES256"] }`,
			expectSess: &AuthenticationSession{
				Subject: "sub",
				Extra:   map[string]interface{}{"sub": "sub", "exp": float64(now.Add(time.Hour).Unix())},
			},
		},
		{
			d: "should pass because JWT is valid",
			r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + generateJWT(t, jwt.MapClaims{
				"sub": "sub",
				"exp": now.Add(time.Hour).Unix(),
			}, "RS256")}}},
			config:    `{}`,
			expectErr: false,
		},
		{
			d: "should fail because JWT nbf is in future",
			r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + generateJWT(t, jwt.MapClaims{
				"sub": "sub",
				"exp": now.Add(time.Hour).Unix(),
				"nbf": now.Add(time.Hour).Unix(),
			}, "RS256")}}},
			config:    `{}`,
			expectErr: true,
		},
		{
			d: "should fail because JWT iat is in future",
			r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + generateJWT(t, jwt.MapClaims{
				"sub": "sub",
				"exp": now.Add(time.Hour).Unix(),
				"iat": now.Add(time.Hour).Unix(),
			}, "RS256")}}},
			config:    `{}`,
			expectErr: true,
		},
		{
			d: "should pass because JWT is missing scope",
			r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + generateJWT(t, jwt.MapClaims{
				"sub":   "sub",
				"exp":   now.Add(time.Hour).Unix(),
				"scope": []string{"scope-1", "scope-2"},
			}, "RS256")}}},
			config:    `{"required_scope": ["scope-1", "scope-2", "scope-3"]}`,
			expectErr: true,
		},
		{
			d: "should pass because JWT issuer is untrusted",
			r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + generateJWT(t, jwt.MapClaims{
				"sub": "sub",
				"exp": now.Add(time.Hour).Unix(),
				"iss": "iss-4",
			}, "RS256")}}},
			config:    `{"trusted_issuers": ["iss-1", "iss-2", "iss-3"]}`,
			expectErr: true,
		},
		{
			d: "should pass because JWT is missing audience",
			r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + generateJWT(t, jwt.MapClaims{
				"sub": "sub",
				"exp": now.Add(time.Hour).Unix(),
				"aud": []string{"aud-1", "aud-2"},
			}, "RS256")}}},
			config:    `{"required_audience": ["aud-1", "aud-2", "aud-3"]}`,
			expectErr: true,
		},
		{
			d: "should fail because JWT is expired",
			r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + generateJWT(t, jwt.MapClaims{
				"sub": "sub",
				"exp": now.Add(-time.Hour).Unix(),
			}, "RS256")}}},
			config:    `{}`,
			expectErr: true,
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
			session, err := authenticator.Authenticate(tc.r, json.RawMessage([]byte(tc.config)), nil)
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
}

func TestGetScopeClaim(t *testing.T) {
	for k, tc := range []struct {
		i map[string]interface{}
		e []string
	}{
		{i: map[string]interface{}{}, e: []string{}},
		{i: map[string]interface{}{"scp": "foo bar"}, e: []string{"foo", "bar"}},
		{i: map[string]interface{}{"scope": "foo bar"}, e: []string{"foo", "bar"}},
		{i: map[string]interface{}{"scopes": "foo bar"}, e: []string{"foo", "bar"}},
		{i: map[string]interface{}{"scp": []string{"foo", "bar"}}, e: []string{"foo", "bar"}},
		{i: map[string]interface{}{"scope": []string{"foo", "bar"}}, e: []string{"foo", "bar"}},
		{i: map[string]interface{}{"scopes": []string{"foo", "bar"}}, e: []string{"foo", "bar"}},
		{i: map[string]interface{}{"scp": []interface{}{"foo", "bar"}}, e: []string{"foo", "bar"}},
		{i: map[string]interface{}{"scope": []interface{}{"foo", "bar"}}, e: []string{"foo", "bar"}},
		{i: map[string]interface{}{"scopes": []interface{}{"foo", "bar"}}, e: []string{"foo", "bar"}},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			assert.EqualValues(t, tc.e, getScopeClaim(tc.i))
		})
	}
}
