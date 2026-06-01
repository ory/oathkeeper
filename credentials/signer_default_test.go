// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package credentials_test

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	. "github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/x"
)

func TestSignerDefault(t *testing.T) {
	reg := internal.NewRegistry(t)
	signer := NewSignerDefault(reg)

	for _, src := range []string{
		"file://../test/stub/jwks-hs.json",
		"file://../test/stub/jwks-rsa-multiple.json",
		"file://../test/stub/jwks-rsa-single.json",
	} {
		t.Run(fmt.Sprintf("src=%s", src), func(t *testing.T) {
			token, err := signer.Sign(t.Context(), x.ParseURLOrPanic(src), jwt.MapClaims{"sub": "foo"})
			require.NoError(t, err)

			fetcher := NewFetcherDefault(reg, time.Second, time.Second)

			_, err = verify(t, token, fetcher, src)
			require.NoError(t, err)
		})
	}

}

func verify(t *testing.T, token string, f Fetcher, u string) (*jwt.Token, error) {
	to, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, errors.New("kid")
		}

		key, err := f.ResolveKey(t.Context(), []url.URL{*x.ParseURLOrPanic(u)}, kid, "sig")
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// transform to public key
		if _, ok := key.Key.([]byte); !ok && !key.IsPublic() {
			key = new(key.Public())
		}

		return key.Key, nil
	})

	return to, err
}
