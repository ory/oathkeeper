// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
)

func TestCookieSessionCache(t *testing.T) {
	t.Parallel()
	logger := logrusx.New("", "")
	c, err := configuration.NewKoanfProvider(
		context.Background(),
		nil,
		logger,
		configx.WithValues(map[string]interface{}{
			"authenticators.cookie_session.config.check_session_url": "http://localhost:8080/",
			"authenticators.cookie_session.config.cache.enabled":     true,
		}))
	require.NoError(t, err)

	a := NewAuthenticatorCookieSession(c, logger, noop.NewTracerProvider())
	assert.Equal(t, "cookie_session", a.GetID())

	config, err := a.Config(nil)
	require.NoError(t, err)

	t.Run("method=sessionToCache", func(t *testing.T) {
		t.Run("case=cache value", func(t *testing.T) {
			req := &http.Request{
				Header: http.Header{},
			}
			req.AddCookie(&http.Cookie{Name: "session", Value: "test-token"})

			subject := "test-subject"
			extra := map[string]interface{}{"role": "admin"}

			a.sessionToCache(config, req, subject, extra)
			time.Sleep(time.Millisecond * 10)

			v := a.sessionFromCache(config, req)
			require.NotNil(t, v)
			require.Equal(t, "test-subject", v.Subject)
			require.Equal(t, map[string]interface{}{"role": "admin"}, v.Extra)
		})

		t.Run("case=value cannot be marshaled to json should not be cached", func(t *testing.T) {
			req := &http.Request{
				Header: http.Header{},
			}
			req.AddCookie(&http.Cookie{Name: "session", Value: "invalid-token"})

			subject := "test-subject"
			extra := map[string]interface{}{"channel": make(chan bool, 1)}

			a.sessionToCache(config, req, subject, extra)
			time.Sleep(time.Millisecond * 10)

			v := a.sessionFromCache(config, req)
			require.Nil(t, v)
		})

		t.Run("case=cached invalid json value should not working", func(t *testing.T) {
			req := &http.Request{
				Header: http.Header{},
			}
			req.AddCookie(&http.Cookie{Name: "session", Value: "bad-json"})

			key := cookiesToCacheKey(req.Cookies())
			ok := a.sessionCache.Set(key, []byte("invalid-json-string"), 1)
			require.True(t, ok)
			time.Sleep(time.Millisecond * 10)

			v := a.sessionFromCache(config, req)
			require.Nil(t, v)
		})

		t.Run("case=cache with ttl", func(t *testing.T) {
			req := &http.Request{
				Header: http.Header{},
			}
			req.AddCookie(&http.Cookie{Name: "session", Value: "ttl-token"})

			subject := "ttl-subject"
			extra := map[string]interface{}{"role": "user"}

			config, _ := a.Config([]byte(`{ "cache": { "ttl": "500ms" } }`))
			a.sessionToCache(config, req, subject, extra)
			a.sessionCache.Wait()

			assert.EventuallyWithT(t, func(t *assert.CollectT) {
				v := a.sessionFromCache(config, req)
				assert.NotNil(t, v)
			}, 490*time.Millisecond, 10*time.Millisecond)

			assert.EventuallyWithT(t, func(t *assert.CollectT) {
				v := a.sessionFromCache(config, req)
				assert.Nil(t, v)
			}, 700*time.Millisecond, 10*time.Millisecond)
		})
	})
}
