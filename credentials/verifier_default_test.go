// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package credentials_test

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

	. "github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/x"
)

func TestVerifierDefault(t *testing.T) {
	reg := internal.NewRegistry(t)

	signer := NewSignerDefault(reg)
	verifier := NewVerifierDefault(reg)
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
		{
			d: "should pass when the trusted issuer has a trailing slash but the token issuer does not (issue #527)",
			c: &ValidationContext{
				Algorithms: []string{"HS256"},
				Issuers:    []string{"https://my-issuer/"},
				KeyURLs:    []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
			},
			token: sign(jwt.MapClaims{
				"sub": "sub",
				"exp": now.Add(time.Hour).Unix(),
				"iss": "https://my-issuer",
			}, "file://../test/stub/jwks-hs.json"),
		},
		{
			d: "should pass when the token issuer has a trailing slash but the trusted issuer does not (issue #527)",
			c: &ValidationContext{
				Algorithms: []string{"HS256"},
				Issuers:    []string{"https://my-issuer"},
				KeyURLs:    []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
			},
			token: sign(jwt.MapClaims{
				"sub": "sub",
				"exp": now.Add(time.Hour).Unix(),
				"iss": "https://my-issuer/",
			}, "file://../test/stub/jwks-hs.json"),
		},
		{
			d: "should still fail when the issuer genuinely differs despite trailing-slash normalization (issue #527)",
			c: &ValidationContext{
				Algorithms: []string{"HS256"},
				Issuers:    []string{"https://my-issuer/"},
				KeyURLs:    []url.URL{*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json")},
			},
			token: sign(jwt.MapClaims{
				"sub": "sub",
				"exp": now.Add(time.Hour).Unix(),
				"iss": "https://evil-issuer",
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
