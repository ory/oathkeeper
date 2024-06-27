// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package credentials_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ory/oathkeeper/internal"
)

func BenchmarkDefaultSigner(b *testing.B) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)
	ctx := context.Background()

	for alg, keys := range map[string]string{
		"RS256": "file://../test/stub/jwks-rsa-multiple.json",
		"ES256": "file://../test/stub/jwks-ecdsa.json",
		"HS256": "file://../test/stub/jwks-hs.json",
	} {
		b.Run("alg="+alg, func(b *testing.B) {
			jwks, _ := url.Parse(keys)
			for i := 0; i < b.N; i++ {
				if _, err := reg.CredentialsSigner().Sign(ctx, jwks, jwt.MapClaims{
					"custom-claim2": 3.14159,
					"custom-claim3": true,
					"exp":           time.Now().Add(time.Minute).Unix(),
					"iat":           time.Now().Unix(),
					"iss":           "issuer",
					"nbf":           time.Now().Unix(),
					"sub":           "some subject",
				}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
