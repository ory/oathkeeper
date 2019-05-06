package credentials

import (
	"context"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/ory/x/urlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
	"time"
)

type defaultSignerMockRegistry struct {
	f Fetcher
}

func newDefaultSignerMockRegistry() *defaultSignerMockRegistry {
	return &defaultSignerMockRegistry{f: NewFetcherDefault(logrus.New(), time.Millisecond*100, time.Millisecond*500)}
}

func (m *defaultSignerMockRegistry) CredentialsFetcher() Fetcher {
	return m.f
}

func TestSignerDefault(t *testing.T) {
	signer := NewDefaultSigner(newDefaultSignerMockRegistry())

	for _, src := range []string{
		"file://../stub/jwks-hs.json",
		"file://../stub/jwks-rsa-multiple.json",
		"file://../stub/jwks-rsa-single.json",
	} {
		t.Run(fmt.Sprintf("src=%s", src), func(t *testing.T) {
			token, err := signer.Sign(context.Background(), urlx.ParseOrPanic(src), jwt.MapClaims{"sub": "foo"})
			require.NoError(t, err)

			fetcher := NewFetcherDefault(logrus.New(), time.Second, time.Second)

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
