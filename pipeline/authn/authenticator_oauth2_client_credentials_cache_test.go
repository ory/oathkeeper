package authn

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/viper"
	"github.com/ory/x/logrusx"
)

func TestClientCredentialsCache(t *testing.T) {
	viper.Reset()

	ts := httptest.NewServer(httprouter.New())

	viper.Set("authenticators.oauth2_client_credentials.config.token_url", ts.URL+"/oauth2/token")
	viper.Set("authenticators.oauth2_client_credentials.config.cache.enabled", true)

	logger := logrusx.New("", "")
	c := configuration.NewViperProvider(logger)
	a := NewAuthenticatorOAuth2ClientCredentials(c, logger)
	assert.Equal(t, "oauth2_client_credentials", a.GetID())

	config, err := a.Config(nil)
	require.NoError(t, err)

	t.Run("method=tokenToCache", func(t *testing.T) {
		t.Run("case=cache value", func(t *testing.T) {
			token := oauth2.Token{
				AccessToken: "some-token",
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(3600 * time.Second),
			}
			cc := clientcredentials.Config{
				ClientID:     "id",
				ClientSecret: "secret",
			}

			a.tokenToCache(config, cc, token)
			// wait for cache to save value
			time.Sleep(time.Millisecond * 10)

			v := a.tokenFromCache(config, cc)
			require.NotNil(t, v)
		})

		t.Run("case=cached invalid json value should not working", func(t *testing.T) {
			cc := clientcredentials.Config{
				ClientID:     "id",
				ClientSecret: "secret",
			}

			ok := a.tokenCache.Set(clientCredentialsConfigToKey(cc), []byte("invalid-json-string"), 1)
			require.True(t, ok)
			// wait cache to save value
			time.Sleep(time.Millisecond * 10)

			v := a.tokenFromCache(config, cc)
			require.Nil(t, v)
		})

		t.Run("case=cache with ttl", func(t *testing.T) {
			token := oauth2.Token{
				AccessToken: "some-token",
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(3600 * time.Second),
			}
			cc := clientcredentials.Config{
				ClientID:     "id",
				ClientSecret: "secret",
			}

			config, _ := a.Config([]byte(`{ "cache": { "ttl": "100ms" } }`))
			a.tokenToCache(config, cc, token)
			// wait cache to save value
			time.Sleep(time.Millisecond * 10)

			v := a.tokenFromCache(config, cc)
			require.NotNil(t, v)

			// wait cache to be expired
			time.Sleep(time.Millisecond * 100)
			v = a.tokenFromCache(config, cc)
			require.Nil(t, v)
		})
	})

}
