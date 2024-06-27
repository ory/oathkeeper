// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package credentials

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/fosite"
	"github.com/ory/oathkeeper/x"
)

func TestVerifierDefault(t *testing.T) {
	signer := NewSignerDefault(newDefaultSignerMockRegistry())
	verifier := NewVerifierDefault(newDefaultSignerMockRegistry())
	now := time.Now().Round(time.Second)

	var sign = func(claims jwt.MapClaims, src string) string {
		tt, err := signer.Sign(context.Background(), x.ParseURLOrPanic(src), claims)
		require.NoError(t, err)
		return tt
	}

	for k, tc := range []struct {
		d            string
		token        string
		c            *ValidationContext
		expectErr    bool
		expectClaims jwt.MapClaims
	}{
		{
			d:         "should fail because JWT is invalid",
			c:         new(ValidationContext),
			expectErr: true,
			token:     "invalid",
		},
		{
			d: "should pass because JWT is valid",
			c: &ValidationContext{
				Algorithms:    []string{"HS256"},
				Audiences:     []string{"aud-1", "aud-2"},
				Issuers:       []string{"iss-1", "iss-2"},
				Scope:         []string{"scope-1", "scope-2"},
				KeyURLs:       []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
				ScopeStrategy: fosite.ExactScopeStrategy,
			},
			token: sign(jwt.MapClaims{
				"sub":   "sub",
				"exp":   now.Add(time.Hour).Unix(),
				"aud":   []string{"aud-1", "aud-2"},
				"iss":   "iss-2",
				"scope": []string{"scope-3", "scope-2", "scope-1"},
			}, "file://../test/stub/jwks-hs.json"),
			expectClaims: jwt.MapClaims{
				"sub": "sub",
				"exp": float64(now.Add(time.Hour).Unix()),
				"aud": []interface{}{"aud-1", "aud-2"},
				"iss": "iss-2",
				"scp": []string{"scope-3", "scope-2", "scope-1"},
			},
		},
		{
			d: "should pass even when scope is a string",
			c: &ValidationContext{
				Algorithms:    []string{"HS256"},
				Audiences:     []string{"aud-1", "aud-2"},
				Issuers:       []string{"iss-1", "iss-2"},
				Scope:         []string{"scope-1", "scope-2"},
				KeyURLs:       []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
				ScopeStrategy: fosite.ExactScopeStrategy,
			},
			token: sign(jwt.MapClaims{
				"sub":   "sub",
				"exp":   now.Add(time.Hour).Unix(),
				"aud":   []string{"aud-1", "aud-2"},
				"iss":   "iss-2",
				"scope": "scope-3 scope-2 scope-1",
			}, "file://../test/stub/jwks-hs.json"),
			expectClaims: jwt.MapClaims{
				"sub": "sub",
				"exp": float64(now.Add(time.Hour).Unix()),
				"aud": []interface{}{"aud-1", "aud-2"},
				"iss": "iss-2",
				"scp": []string{"scope-3", "scope-2", "scope-1"},
			},
		},
		{
			d: "should pass when scope is keyed as scp",
			c: &ValidationContext{
				Algorithms:    []string{"HS256"},
				Audiences:     []string{"aud-1", "aud-2"},
				Issuers:       []string{"iss-1", "iss-2"},
				Scope:         []string{"scope-1", "scope-2"},
				KeyURLs:       []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
				ScopeStrategy: fosite.ExactScopeStrategy,
			},
			token: sign(jwt.MapClaims{
				"sub": "sub",
				"exp": now.Add(time.Hour).Unix(),
				"aud": []string{"aud-1", "aud-2"},
				"iss": "iss-2",
				"scp": "scope-3 scope-2 scope-1",
			}, "file://../test/stub/jwks-hs.json"),
			expectClaims: jwt.MapClaims{
				"sub": "sub",
				"exp": float64(now.Add(time.Hour).Unix()),
				"aud": []interface{}{"aud-1", "aud-2"},
				"iss": "iss-2",
				"scp": []string{"scope-3", "scope-2", "scope-1"},
			},
		},
		{
			d: "should pass when scope is keyed as scopes",
			c: &ValidationContext{
				Algorithms:    []string{"HS256"},
				Audiences:     []string{"aud-1", "aud-2"},
				Issuers:       []string{"iss-1", "iss-2"},
				Scope:         []string{"scope-1", "scope-2"},
				KeyURLs:       []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
				ScopeStrategy: fosite.ExactScopeStrategy,
			},
			token: sign(jwt.MapClaims{
				"sub":    "sub",
				"exp":    now.Add(time.Hour).Unix(),
				"aud":    []string{"aud-1", "aud-2"},
				"iss":    "iss-2",
				"scopes": "scope-3 scope-2 scope-1",
			}, "file://../test/stub/jwks-hs.json"),
			expectClaims: jwt.MapClaims{
				"sub": "sub",
				"exp": float64(now.Add(time.Hour).Unix()),
				"aud": []interface{}{"aud-1", "aud-2"},
				"iss": "iss-2",
				"scp": []string{"scope-3", "scope-2", "scope-1"},
			},
		},
		{
			d: "should fail when scope validation was requested but no scope strategy is set",
			c: &ValidationContext{
				Algorithms: []string{"HS256"},
				Audiences:  []string{"aud-1", "aud-2"},
				Issuers:    []string{"iss-1", "iss-2"},
				Scope:      []string{"scope-1", "scope-2"},
				KeyURLs:    []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
			},
			token: sign(jwt.MapClaims{
				"sub":    "sub",
				"exp":    now.Add(time.Hour).Unix(),
				"aud":    []string{"aud-1", "aud-2"},
				"iss":    "iss-2",
				"scopes": "scope-3 scope-2 scope-1",
			}, "file://../test/stub/jwks-hs.json"),
			expectErr: true,
		},
		{
			d: "should fail when algorithm does not match",
			c: &ValidationContext{
				Algorithms:    []string{"HS256"},
				Audiences:     []string{"aud-1", "aud-2"},
				Issuers:       []string{"iss-1", "iss-2"},
				Scope:         []string{"scope-1", "scope-2"},
				KeyURLs:       []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-rsa-single.json")},
				ScopeStrategy: fosite.ExactScopeStrategy,
			},
			token: sign(jwt.MapClaims{
				"sub":   "sub",
				"exp":   now.Add(time.Hour).Unix(),
				"aud":   []string{"aud-1", "aud-2"},
				"iss":   "iss-2",
				"scope": "scope-3 scope-2 scope-1",
			}, "file://../test/stub/jwks-rsa-single.json"),
			expectErr: true,
		},
		{
			d: "should fail when audience mismatches",
			c: &ValidationContext{
				Algorithms:    []string{"HS256"},
				Audiences:     []string{"aud-1", "aud-2"},
				Issuers:       []string{"iss-1", "iss-2"},
				Scope:         []string{"scope-1", "scope-2"},
				KeyURLs:       []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
				ScopeStrategy: fosite.ExactScopeStrategy,
			},
			token: sign(jwt.MapClaims{
				"sub":   "sub",
				"exp":   now.Add(time.Hour).Unix(),
				"aud":   []string{"not-aud-1", "aud-2"},
				"iss":   "iss-2",
				"scope": "scope-3 scope-2 scope-1",
			}, "file://../test/stub/jwks-hs.json"),
			expectErr: true,
		},
		{
			d: "should fail when issuer mismatches",
			c: &ValidationContext{
				Algorithms:    []string{"HS256"},
				Audiences:     []string{"aud-1", "aud-2"},
				Issuers:       []string{"iss-1", "iss-2"},
				Scope:         []string{"scope-1", "scope-2"},
				KeyURLs:       []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
				ScopeStrategy: fosite.ExactScopeStrategy,
			},
			token: sign(jwt.MapClaims{
				"sub":   "sub",
				"exp":   now.Add(time.Hour).Unix(),
				"aud":   []string{"aud-1", "aud-2"},
				"iss":   "not-iss-2",
				"scope": "scope-3 scope-2 scope-1",
			}, "file://../test/stub/jwks-hs.json"),
			expectErr: true,
		},
		{
			d: "should fail when issuer mismatches",
			c: &ValidationContext{
				Algorithms:    []string{"HS256"},
				Audiences:     []string{"aud-1", "aud-2"},
				Issuers:       []string{"iss-1", "iss-2"},
				Scope:         []string{"scope-1", "scope-2"},
				KeyURLs:       []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
				ScopeStrategy: fosite.ExactScopeStrategy,
			},
			token: sign(jwt.MapClaims{
				"sub":   "sub",
				"exp":   now.Add(time.Hour).Unix(),
				"aud":   []string{"aud-1", "aud-2"},
				"iss":   "iss-2",
				"scope": "scope-3 not-scope-2 scope-1",
			}, "file://../test/stub/jwks-hs.json"),
			expectErr: true,
		},
		{
			d: "should fail when expired",
			c: &ValidationContext{
				Algorithms:    []string{"HS256"},
				Audiences:     []string{"aud-1", "aud-2"},
				Issuers:       []string{"iss-1", "iss-2"},
				Scope:         []string{"scope-1", "scope-2"},
				KeyURLs:       []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
				ScopeStrategy: fosite.ExactScopeStrategy,
			},
			token: sign(jwt.MapClaims{
				"sub":   "sub",
				"exp":   now.Add(-time.Hour).Unix(),
				"aud":   []string{"aud-1", "aud-2"},
				"iss":   "iss-2",
				"scope": "scope-3 scope-2 scope-1",
			}, "file://../test/stub/jwks-hs.json"),
			expectErr: true,
		},
		{
			d: "should fail when nbf in future",
			c: &ValidationContext{
				Algorithms:    []string{"HS256"},
				Audiences:     []string{"aud-1", "aud-2"},
				Issuers:       []string{"iss-1", "iss-2"},
				Scope:         []string{"scope-1", "scope-2"},
				KeyURLs:       []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
				ScopeStrategy: fosite.ExactScopeStrategy,
			},
			token: sign(jwt.MapClaims{
				"sub":   "sub",
				"exp":   now.Add(time.Hour).Unix(),
				"nbf":   now.Add(time.Hour).Unix(),
				"aud":   []string{"aud-1", "aud-2"},
				"iss":   "iss-2",
				"scope": "scope-3 scope-2 scope-1",
			}, "file://../test/stub/jwks-hs.json"),
			expectErr: true,
		},
		{
			d: "should fail when iat in future",
			c: &ValidationContext{
				Algorithms:    []string{"HS256"},
				Audiences:     []string{"aud-1", "aud-2"},
				Issuers:       []string{"iss-1", "iss-2"},
				Scope:         []string{"scope-1", "scope-2"},
				KeyURLs:       []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
				ScopeStrategy: fosite.ExactScopeStrategy,
			},
			token: sign(jwt.MapClaims{
				"sub":   "sub",
				"exp":   now.Add(time.Hour).Unix(),
				"iat":   now.Add(time.Hour).Unix(),
				"aud":   []string{"aud-1", "aud-2"},
				"iss":   "iss-2",
				"scope": "scope-3 scope-2 scope-1",
			}, "file://../test/stub/jwks-hs.json"),
			expectErr: true,
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
			claims, err := verifier.Verify(context.Background(), tc.token, tc.c)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err, "%+v", errors.Cause(err), err)
			if tc.expectClaims != nil {
				assert.EqualValues(t, tc.expectClaims, claims.Claims)
			}
		})
	}
}

func TestScope(t *testing.T) {
	for k, tc := range []struct {
		i  map[string]interface{}
		ev []string
		ek string
	}{
		{i: map[string]interface{}{}, ev: []string{}},
		{i: map[string]interface{}{"scp": "foo bar"}, ev: []string{"foo", "bar"}, ek: "scp"},
		{i: map[string]interface{}{"scope": "foo bar"}, ev: []string{"foo", "bar"}, ek: "scope"},
		{i: map[string]interface{}{"scopes": "foo bar"}, ev: []string{"foo", "bar"}, ek: "scopes"},
		{i: map[string]interface{}{"scp": []string{"foo", "bar"}}, ev: []string{"foo", "bar"}, ek: "scp"},
		{i: map[string]interface{}{"scope": []string{"foo", "bar"}}, ev: []string{"foo", "bar"}, ek: "scope"},
		{i: map[string]interface{}{"scopes": []string{"foo", "bar"}}, ev: []string{"foo", "bar"}, ek: "scopes"},
		{i: map[string]interface{}{"scp": []interface{}{"foo", "bar"}}, ev: []string{"foo", "bar"}, ek: "scp"},
		{i: map[string]interface{}{"scope": []interface{}{"foo", "bar"}}, ev: []string{"foo", "bar"}, ek: "scope"},
		{i: map[string]interface{}{"scopes": []interface{}{"foo", "bar"}}, ev: []string{"foo", "bar"}, ek: "scopes"},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			value, key := scope(tc.i)
			assert.EqualValues(t, tc.ev, value)
			assert.EqualValues(t, tc.ek, key)
		})
	}
}
