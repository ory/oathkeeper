package credentials_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
)

func BenchmarkDefaultSigner(b *testing.B) {
	conf, err := configuration.NewViperProvider(context.Background(), logrusx.New("", ""),
		configx.WithValue("log.level", "debug"),
		configx.WithValue(configuration.ViperKeyErrorsJSONIsEnabled, true))
	require.NoError(b, err)

	reg := driver.NewRegistryMemory().WithConfig(conf)
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
