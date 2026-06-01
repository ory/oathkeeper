// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/fosite"
	"github.com/ory/x/configx"
	"github.com/ory/x/uuidx"

	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/pipeline/authn"
)

func TestCache(t *testing.T) {
	t.Parallel()
	reg := internal.NewRegistry(t, configx.WithValues(map[string]interface{}{
		"authenticators.oauth2_introspection.config.cache.enabled":     true,
		"authenticators.oauth2_introspection.config.introspection_url": "http://localhost:8080/",
	}))

	a := authn.NewAuthenticatorOAuth2Introspection(reg)
	assert.Equal(t, "oauth2_introspection", a.GetID())

	config, _, err := a.Config(nil)
	require.NoError(t, err)

	t.Run("case=cache value", func(t *testing.T) {
		i := &authn.AuthenticatorOAuth2IntrospectionResult{
			Active: true,
			Extra:  map[string]interface{}{"extra": "foo"},
		}

		a.TokenToCache(config, i, "token", fosite.WildcardScopeStrategy)
		a.WaitForCache()

		// modify struct should not affect cached value
		i.Active = false
		v := a.TokenFromCache(config, "token", fosite.WildcardScopeStrategy)
		require.NotNil(t, v)
		require.True(t, v.Active)
	})

	t.Run("case=value cannot be marshaled to json should not be cached", func(t *testing.T) {
		i := &authn.AuthenticatorOAuth2IntrospectionResult{
			Active: false,
			Extra:  map[string]interface{}{"extra": make(chan bool, 1)},
		}

		a.TokenToCache(config, i, "invalid-token", fosite.WildcardScopeStrategy)
		a.WaitForCache()

		v := a.TokenFromCache(config, "invalid-token", fosite.WildcardScopeStrategy)
		require.Nil(t, v)
	})

	t.Run("case=cached invalid json", func(t *testing.T) {
		ok := a.TokenCache.Set(authn.TokenCacheKey("invalid-json", config.IntrospectionURL), []byte("invalid-json-string"), 1)
		require.True(t, ok)
		a.WaitForCache()

		v := a.TokenFromCache(config, "invalid-json", fosite.WildcardScopeStrategy)
		require.Nil(t, v)
	})

	t.Run("case=cache with ttl", func(t *testing.T) {
		i := &authn.AuthenticatorOAuth2IntrospectionResult{Active: true}

		config, _, err := a.Config([]byte(`{ "cache": { "ttl": "500ms" } }`))
		require.NoError(t, err)
		a.TokenToCache(config, i, "token", fosite.WildcardScopeStrategy)
		a.TokenCache.Wait()

		assert.EventuallyWithT(t, func(t *assert.CollectT) {
			v := a.TokenFromCache(config, "token", fosite.WildcardScopeStrategy)
			assert.NotNil(t, v)
		}, 490*time.Millisecond, 10*time.Millisecond)

		// wait cache to be expired
		assert.EventuallyWithT(t, func(t *assert.CollectT) {
			v := a.TokenFromCache(config, "token", fosite.WildcardScopeStrategy)
			assert.Nil(t, v)
		}, 700*time.Millisecond, 10*time.Millisecond)
	})

	t.Run("case=token with different introspection URL", func(t *testing.T) {
		i := &authn.AuthenticatorOAuth2IntrospectionResult{Active: true}

		config, _, err := a.Config([]byte(`{ "cache": { "ttl": "0s" }, "introspection_url": "http://localhost/oauth2/token" }`))
		require.NoError(t, err)

		token := uuidx.NewV4().String()
		a.TokenToCache(config, i, token, fosite.WildcardScopeStrategy)
		a.WaitForCache()

		config, _, err = a.Config([]byte(`{ "cache": { "ttl": "0s" }, "introspection_url": "http://localhost/oauth2/token2" }`))
		require.NoError(t, err)

		v := a.TokenFromCache(config, token, fosite.WildcardScopeStrategy)
		require.Nil(t, v)
	})
}
