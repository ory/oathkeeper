// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/x/configx"

	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/pipeline/authn"
)

func TestClientCredentialsCache(t *testing.T) {
	t.Parallel()

	reg := internal.NewRegistry(t, configx.WithValues(map[string]interface{}{
		"authenticators.oauth2_client_credentials.config.token_url":     "https://example.com/oauth2/token",
		"authenticators.oauth2_client_credentials.config.cache.enabled": true,
	}))

	a := authn.NewAuthenticatorOAuth2ClientCredentials(reg)
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

			a.TokenToCache(config, cc, token)
			// wait for cache to save value
			time.Sleep(time.Millisecond * 10)

			v := a.TokenFromCache(config, cc)
			require.NotNil(t, v)
		})

		t.Run("case=cached invalid json value should not working", func(t *testing.T) {
			cc := clientcredentials.Config{
				ClientID:     "id",
				ClientSecret: "secret",
			}

			ok := a.TokenCache.Set(authn.ClientCredentialsConfigToKey(cc), []byte("invalid-json-string"), 1)
			require.True(t, ok)
			// wait cache to save value
			time.Sleep(time.Millisecond * 10)

			v := a.TokenFromCache(config, cc)
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
			a.TokenToCache(config, cc, token)
			a.TokenCache.Wait()

			assert.EventuallyWithT(t, func(t *assert.CollectT) {
				v := a.TokenFromCache(config, cc)
				assert.NotNil(t, v)
			}, 90*time.Millisecond, 10*time.Millisecond)

			// wait cache to be expired
			assert.EventuallyWithT(t, func(t *assert.CollectT) {
				v := a.TokenFromCache(config, cc)
				assert.Nil(t, v)
			}, 120*time.Millisecond, 10*time.Millisecond)
		})
	})
}
