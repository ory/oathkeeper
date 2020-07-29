package credentials

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/logrusx"

	"github.com/ory/x/urlx"
)

type defaultSignerMockRegistry struct {
	f Fetcher
}

func newDefaultSignerMockRegistry() *defaultSignerMockRegistry {
	return &defaultSignerMockRegistry{f: NewFetcherDefault(logrusx.New("", ""), time.Millisecond*100, time.Millisecond*500)}
}

func (m *defaultSignerMockRegistry) CredentialsFetcher() Fetcher {
	return m.f
}

func TestSignerDefault(t *testing.T) {
	signer := NewSignerDefault(newDefaultSignerMockRegistry())

	for _, src := range []string{
		"file://../test/stub/jwks-hs.json",
		"file://../test/stub/jwks-rsa-multiple.json",
		"file://../test/stub/jwks-rsa-single.json",
	} {
		t.Run(fmt.Sprintf("src=%s", src), func(t *testing.T) {
			token, err := signer.Sign(context.Background(), urlx.ParseOrPanic(src), jwt.MapClaims{"sub": "foo"})
			require.NoError(t, err)

			fetcher := NewFetcherDefault(logrusx.New("", ""), time.Second, time.Second)

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

		t.Logf("Looking up kid: %s", kid)

		key, err := f.ResolveKey(context.Background(), []url.URL{*urlx.ParseOrPanic(u)}, kid, "sig")
		if err != nil {
			t.Logf("erri erro: %+v", err)
			return nil, errors.WithStack(err)
		}

		// transform to public key
		if _, ok := key.Key.([]byte); !ok && !key.IsPublic() {
			k := key.Public()
			key = &k
		}

		t.Logf("erri erro: %T", key.Key)
		return key.Key, nil
	})

	t.Logf("erri erro: %+v", err)
	return to, err
}
