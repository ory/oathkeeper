//go:build race
// +build race

// Package authn_test contains a concurrent regression harness for the OAuth2
// introspection authenticator. Run with:
//
//	go test -race -v ./pipeline/authn -run TestConfigDataRace
//
// After the stateless-after-init refactoring, Config() no longer mutates
// a.TokenCache or a.cacheTTL (the latter has been removed from the struct).
// TokenToCache and TokenFromCache derive all state from their config argument
// and from the immutable TokenCache pointer set in the constructor, so they
// require no locks and present no shared mutable state to the race detector.
// This test serves as a permanent regression guard: it must pass cleanly
// under -race in all future states of the codebase.
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
// cache TTLs while readers concurrently exercise TokenToCache/TokenFromCache.
// With the stateless-after-init design the Go race detector must not report any
// data races: Config() only writes to clientMap (protected by a.mu.RWMutex),
// and the cache methods access only the immutable TokenCache pointer and the
// request-scoped config argument.
func TestConfigDataRace(t *testing.T) {
	reg := internal.NewRegistry(t, configx.SkipValidation())
	aa, err := reg.PipelineAuthenticator("oauth2_introspection")
	require.NoError(t, err)
	a, ok := aa.(*AuthenticatorOAuth2Introspection)
	require.Truef(t, ok, "got type %T", aa)

	base := `{"introspection_url":"http://127.0.0.1/oauth2/introspect","cache":{"enabled":true,"max_cost":1000}}`
	ttls := []string{"25ms", "50ms", "75ms"}

	var wg sync.WaitGroup

	// Writers: call Config() concurrently with varying TTLs to exercise the
	// clientMap read/write path under a.mu.RWMutex.
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

	// Readers: call Config() then exercise the cache paths concurrently with
	// the writer goroutines. No struct-level mutable state is accessed.
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
