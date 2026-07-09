//go:build race
// +build race

// This test is a minimal concurrent harness to reproduce the data races
// in AuthenticatorOAuth2Introspection.Config() vs TokenToCache/TokenFromCache.
// Run with: go test -race -v ./pipeline/authn -run TestConfigDataRace
package authn_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/x/configx"

	. "github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/internal"
)

// TestConfigDataRace launches concurrent writers invoking Config() with varying
// cache TTLs while readers concurrently hit TokenToCache/TokenFromCache.
// The Go race detector should report races on a.cacheTTL and a.TokenCache.
func TestConfigDataRace(t *testing.T) {
	reg := internal.NewRegistry(t, configx.SkipValidation())
	aa, err := reg.PipelineAuthenticator("oauth2_introspection")
	require.NoError(t, err)
	a, ok := aa.(*AuthenticatorOAuth2Introspection)
	require.Truef(t, ok, "got type %T", aa)

	base := `{"introspection_url":"http://127.0.0.1/oauth2/introspect","cache":{"enabled":true,"max_cost":1000}}`
	ttls := []string{"25ms", "50ms", "75ms"}

	var wg sync.WaitGroup

	// Writers mutate TTL and (re)initialize cache concurrently.
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 2000; j++ {
				cfg := fmt.Sprintf(`{"introspection_url":"http://127.0.0.1/oauth2/introspect","cache":{"enabled":true,"max_cost":1000,"ttl":"%s"}}`, ttls[j%len(ttls)])
				_, _, _ = a.Config([]byte(cfg))
			}
		}(i)
	}

	// Readers exercise cache paths concurrently with Config() calls.
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			res := &AuthenticatorOAuth2IntrospectionResult{Active: true}
			for j := 0; j < 2000; j++ {
				c, _, err := a.Config([]byte(base))
				if err != nil || c == nil {
					continue
				}
				a.TokenToCache(c, res, "tok", nil)
				_ = a.TokenFromCache(c, "tok", nil)
			}
		}(i)
	}

	wg.Wait()
	a.WaitForCache()
}
