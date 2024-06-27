// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/internal"
	. "github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/x"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticatorJWT(t *testing.T) {
	t.Parallel()
	keys := []string{
		"file://../../test/stub/jwks-hs.json",
		"file://../../test/stub/jwks-rsa-multiple.json",
		"file://../../test/stub/jwks-rsa-single.json",
		"file://../../test/stub/jwks-ecdsa.json",
	}
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineAuthenticator("jwt")
	require.NoError(t, err)
	assert.Equal(t, "jwt", a.GetID())

	var gen = func(l string, c jwt.Claims) string {
		token, err := reg.CredentialsSigner().Sign(context.Background(), x.ParseURLOrPanic(l), c)
		require.NoError(t, err)
		return token
	}

	now := time.Now().UTC()

	t.Run("method=authenticate", func(t *testing.T) {
		for k, tc := range []struct {
			setup          func()
			d              string
			r              *http.Request
			config         string
			expectErr      bool
			expectExactErr error
			expectCode     int
			expectSess     *AuthenticationSession
			extraErrAssert func(err error)
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
				d: "should return error saying that authenticator is not responsible for validating the request, as the token was not provided in a proper location (default)",
				r: &http.Request{Header: http.Header{"Foobar": []string{"bearer " + gen(keys[1], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
				})}}},
				expectErr:      true,
				expectExactErr: ErrAuthenticatorNotResponsible,
			},
			{
				d: "should return error saying that authenticator is not responsible for validating the request, as the token was not provided in a proper location (custom header)",
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[1], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
				})}}},
				config:         `{"token_from": {"header": "X-Custom-Header"}}`,
				expectErr:      true,
				expectExactErr: ErrAuthenticatorNotResponsible,
			},
			{
				d: "should return error saying that authenticator is not responsible for validating the request, as the token was not provided in a proper location (custom query parameter)",
				r: &http.Request{
					Form: map[string][]string{
						"someOtherQueryParam": {
							gen(keys[1], jwt.MapClaims{
								"sub": "sub",
								"exp": now.Add(time.Hour).Unix(),
							}),
						},
					},
					Header: http.Header{"Authorization": []string{"bearer " + gen(keys[1], jwt.MapClaims{
						"sub": "sub",
						"exp": now.Add(time.Hour).Unix(),
					})}},
				},
				config:         `{"token_from": {"query_parameter": "token"}}`,
				expectErr:      true,
				expectExactErr: ErrAuthenticatorNotResponsible,
			},
			{
				d: "should return error saying that authenticator is not responsible for validating the request, as the token was not provided in a proper location (cookie)",
				r: &http.Request{Header: http.Header{"Cookie": []string{"biscuit=" + gen(keys[1], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
				})}}},
				config:         `{"token_from": {"cookie": "cake"}}`,
				expectErr:      true,
				expectExactErr: ErrAuthenticatorNotResponsible,
			},
			{
				d:              "should return error saying that authenticator is not responsible for validating the request, as the token provided is not a JWT",
				r:              &http.Request{Header: http.Header{"X-Custom-Header": []string{"bm90LWp3dA=="}}}, // not-jwt
				config:         `{"token_from": {"header": "X-Custom-Header"}}`,
				expectErr:      true,
				expectExactErr: ErrAuthenticatorNotResponsible,
			},
			{
				d: "should pass because the valid JWT token was provided in a proper location (custom header)",
				r: &http.Request{Header: http.Header{"X-Custom-Header": []string{gen(keys[1], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
				})}}},
				config:    `{"token_from": {"header": "X-Custom-Header"}}`,
				expectErr: false,
			},
			{
				d: "should pass because the valid JWT token was provided in a proper location (custom query parameter)",
				r: &http.Request{
					URL: &url.URL{
						RawQuery: "token=" + gen(keys[1], jwt.MapClaims{
							"sub": "sub",
							"exp": now.Add(time.Hour).Unix(),
						}),
					},
				},
				config:    `{"token_from": {"query_parameter": "token"}}`,
				expectErr: false,
			},
			{
				d: "should pass because the valid JWT token was provided in a proper location (cookie)",
				r: &http.Request{Header: http.Header{"Cookie": []string{"biscuit=" + gen(keys[1], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
				})}}},
				config:    `{"token_from": {"cookie": "biscuit"}}`,
				expectErr: false,
			},
			{
				d: "should pass because JWT is valid",
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[1], jwt.MapClaims{
					"sub":   "sub",
					"exp":   now.Add(time.Hour).Unix(),
					"aud":   []string{"aud-1", "aud-2"},
					"iss":   "iss-2",
					"scope": []string{"scope-3", "scope-2", "scope-1"},
				})}}},
				config:    `{"target_audience": ["aud-1", "aud-2"], "trusted_issuers": ["iss-1", "iss-2"], "required_scope": ["scope-1", "scope-2"], "scope_strategy":"exact"}`,
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
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[2], jwt.MapClaims{
					"sub":   "sub",
					"exp":   now.Add(time.Hour).Unix(),
					"aud":   []string{"aud-1", "aud-2"},
					"iss":   "iss-2",
					"scope": "scope-3 scope-2 scope-1",
				})}}},
				config:    `{"target_audience": ["aud-1", "aud-2"], "trusted_issuers": ["iss-1", "iss-2"], "required_scope": ["scope-1", "scope-2"], "scope_strategy":"exact"}`,
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
				config:     `{}`,
				expectErr:  true,
				expectCode: 401,
			},
			{
				d: "should fail because JWT iat is in future",
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[2], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
					"iat": now.Add(time.Hour).Unix(),
				})}}},
				config:     `{}`,
				expectErr:  true,
				expectCode: 401,
			},
			{
				d: "should fail because JWT is missing scope",
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[2], jwt.MapClaims{
					"sub":   "sub",
					"exp":   now.Add(time.Hour).Unix(),
					"scope": []string{"scope-1", "scope-2"},
				})}}},
				config:    `{"required_scope": ["scope-1", "scope-2", "scope-3"]}`,
				expectErr: true,
			},
			{
				d: "should fail because JWT issuer is untrusted",
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[1], jwt.MapClaims{
					"sub": "sub",
					"exp": now.Add(time.Hour).Unix(),
					"iss": "iss-4",
				})}}},
				config:    `{"trusted_issuers": ["iss-1", "iss-2", "iss-3"]}`,
				expectErr: true,
			},
			{
				d: "should fail because JWT is missing audience",
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
				config:     `{}`,
				expectErr:  true,
				expectCode: 401,
			},
			{
				d: "failed JWT authorization results in error with jwt_claims in DetailsField",
				r: &http.Request{Header: http.Header{"Authorization": []string{"bearer " + gen(keys[2], jwt.MapClaims{
					"sub":   "sub",
					"exp":   now.Add(time.Hour).Unix(),
					"scope": []string{"scope-1", "scope-2"},
					"realm_access": map[string][]string{
						"roles": {
							"role-1",
							"role-2",
						},
					},
				})}}},
				config:    `{"required_scope": ["scope-1", "scope-2", "scope-3"]}`,
				expectErr: true,
				extraErrAssert: func(err error) {
					defaultError := err.(*herodot.DefaultError)
					require.Error(t, defaultError)
					require.Equal(t, fmt.Sprintf("{\"exp\":%v,\"realm_access\":{\"roles\":[\"role-1\",\"role-2\"]},\"scope\":[\"scope-1\",\"scope-2\"],\"sub\":\"sub\"}", now.Add(time.Hour).Unix()), defaultError.DetailsField["jwt_claims"])
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
				if tc.setup != nil {
					tc.setup()
				}

				tc.config, _ = sjson.Set(tc.config, "jwks_urls", keys)
				session := new(AuthenticationSession)
				err := a.Authenticate(tc.r, session, json.RawMessage([]byte(tc.config)), nil)
				if tc.expectErr {
					require.Error(t, err)
					if tc.expectCode != 0 {
						assert.Equal(t, tc.expectCode, herodot.ToDefaultError(err, "").StatusCode(), "Status code mismatch")
					}
					if tc.expectExactErr != nil {
						assert.EqualError(t, err, tc.expectExactErr.Error())
					}
					if tc.extraErrAssert != nil {
						tc.extraErrAssert(err)
					}
				} else {
					require.NoError(t, err, "%#v", errors.Cause(err))
				}

				if tc.expectSess != nil {
					assert.Equal(t, tc.expectSess, session)
				}
			})
		}
	})
}
