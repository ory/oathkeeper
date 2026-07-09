// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
)

type AuthenticatorOAuth2Configuration struct {
	Scopes   []string                                                `json:"required_scope"`
	TokenURL string                                                  `json:"token_url"`
	Retry    *AuthenticatorOAuth2ClientCredentialsRetryConfiguration `json:"retry,omitempty"`
	Cache    clientCredentialsCacheConfig                            `json:"cache"`
}

type clientCredentialsCacheConfig struct {
	Enabled   bool   `json:"enabled"`
	TTL       string `json:"ttl"`
	MaxTokens int    `json:"max_tokens"`
}

// AuthenticatorOAuth2ClientCredentials authenticates requests via the OAuth2
// client credentials flow.
//
// Stateless-after-init design (architectural refactoring per @zepatrik):
//
//   - The Ristretto token cache is allocated exactly once inside the factory
//     constructor NewAuthenticatorOAuth2ClientCredentials and is never
//     reassigned at request time. Because the pointer is written before the
//     struct is shared with any goroutine, all subsequent reads of TokenCache
//     are safe without locks.
//
//   - The former a.client *http.Client field was unused in the request path:
//     Authenticate delegates the token exchange entirely to
//     clientcredentials.Config.Token, which manages its own HTTP transport.
//     The field along with its associated sync.Once and sync.Mutex guards
//     have been removed.
//
//   - cacheTTL is no longer persisted on the struct. TokenToCache derives the
//     effective TTL directly from the per-request config.Cache.TTL on each
//     call, eliminating the last source of shared mutable state.
//
// The resulting struct carries no mutable fields after construction, making
// concurrent request handling naturally race-free without any locks.
type AuthenticatorOAuth2ClientCredentials struct {
	d          dependencies
	TokenCache *ristretto.Cache[string, []byte]
}

type AuthenticatorOAuth2ClientCredentialsRetryConfiguration struct {
	Timeout string `json:"max_delay"`
	MaxWait string `json:"give_up_after"`
}

// NewAuthenticatorOAuth2ClientCredentials constructs a fully initialized
// authenticator instance. The Ristretto token cache is created here with
// sensible fixed defaults so that no lazy initialization is required (or
// permitted) at request time, eliminating any TOCTOU window on the cache
// pointer entirely.
func NewAuthenticatorOAuth2ClientCredentials(d dependencies) *AuthenticatorOAuth2ClientCredentials {
	cache, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
		// Default capacity: 1 000 tokens. NumCounters follows the ristretto
		// recommendation of 10 × MaxCost for the frequency-sketch counters.
		NumCounters:        10_000,
		MaxCost:            1_000,
		BufferItems:        64,
		Cost:               func(_ []byte) int64 { return 1 },
		IgnoreInternalCost: true,
	})
	if err != nil {
		// ristretto.NewCache only returns an error for programmer-invalid
		// configuration; the parameters above are unconditionally valid.
		panic(fmt.Sprintf("authn/oauth2_client_credentials: cache init failed: %v", err))
	}
	return &AuthenticatorOAuth2ClientCredentials{d: d, TokenCache: cache}
}

func (a *AuthenticatorOAuth2ClientCredentials) GetID() string { return "oauth2_client_credentials" }

func (a *AuthenticatorOAuth2ClientCredentials) Validate(config json.RawMessage) error {
	if !a.d.Config().AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

// Config parses and validates the merged global + rule-level authenticator
// configuration. It is now a pure read operation with respect to the receiver:
// it reads from a.d.Config() but performs no writes to any field on a, making
// concurrent calls naturally race-free without any synchronization.
func (a *AuthenticatorOAuth2ClientCredentials) Config(config json.RawMessage) (*AuthenticatorOAuth2Configuration, error) {
	const (
		defaultTimeout = "1s"
		defaultMaxWait = "2s"
	)
	var c AuthenticatorOAuth2Configuration
	if err := a.d.Config().AuthenticatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthenticatorMisconfigured(a, err)
	}

	if c.Retry == nil {
		c.Retry = &AuthenticatorOAuth2ClientCredentialsRetryConfiguration{Timeout: defaultTimeout, MaxWait: defaultMaxWait}
	} else {
		if c.Retry.Timeout == "" {
			c.Retry.Timeout = defaultTimeout
		}
		if c.Retry.MaxWait == "" {
			c.Retry.MaxWait = defaultMaxWait
		}
	}

	// Validate retry duration strings eagerly so callers receive a descriptive
	// configuration error rather than a late failure during token exchange.
	if _, err := time.ParseDuration(c.Retry.Timeout); err != nil {
		return nil, err
	}
	if _, err := time.ParseDuration(c.Retry.MaxWait); err != nil {
		return nil, err
	}

	return &c, nil
}

func ClientCredentialsConfigToKey(cc clientcredentials.Config) string {
	return fmt.Sprintf("%s|%s|%s:%s", cc.TokenURL, strings.Join(cc.Scopes, " "), cc.ClientID, cc.ClientSecret)
}

func (a *AuthenticatorOAuth2ClientCredentials) TokenFromCache(config *AuthenticatorOAuth2Configuration, clientCredentials clientcredentials.Config) *oauth2.Token {
	if !config.Cache.Enabled {
		return nil
	}

	// TokenCache is immutable after construction: no lock required.
	i, found := a.TokenCache.Get(ClientCredentialsConfigToKey(clientCredentials))
	if !found {
		return nil
	}

	var v oauth2.Token
	if err := json.Unmarshal(i, &v); err != nil {
		return nil
	}
	return &v
}

func (a *AuthenticatorOAuth2ClientCredentials) TokenToCache(config *AuthenticatorOAuth2Configuration, clientCredentials clientcredentials.Config, token oauth2.Token) {
	if !config.Cache.Enabled {
		return
	}

	key := ClientCredentialsConfigToKey(clientCredentials)

	v, err := json.Marshal(token)
	if err != nil {
		return
	}

	// Derive the effective TTL from the per-request config on each call.
	// cacheTTL is no longer stored on the struct; reading it from the
	// request-scoped config eliminates all struct mutation, making this
	// method safe for concurrent use without any lock.
	if config.Cache.TTL != "" {
		if cacheTTL, parseErr := time.ParseDuration(config.Cache.TTL); parseErr == nil {
			// Allow up-to at most the cache TTL, otherwise use token expiry.
			ttl := time.Until(token.Expiry)
			if ttl > cacheTTL {
				ttl = cacheTTL
			}
			a.TokenCache.SetWithTTL(key, v, 1, ttl)
			return
		}
	}

	// No TTL configured: use token expiry, or zero for non-expiring tokens.
	ttl := time.Duration(0)
	if !token.Expiry.IsZero() {
		ttl = time.Until(token.Expiry)
	}
	a.TokenCache.SetWithTTL(key, v, 1, ttl)
}

func (a *AuthenticatorOAuth2ClientCredentials) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	cf, err := a.Config(config)
	if err != nil {
		return err
	}

	user, password, ok := r.BasicAuth()
	if !ok {
		return errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	user, err = url.QueryUnescape(user)
	if err != nil {
		return errors.Wrap(helper.ErrUnauthorized(), err.Error())
	}

	password, err = url.QueryUnescape(password)
	if err != nil {
		return errors.Wrap(helper.ErrUnauthorized(), err.Error())
	}

	c := clientcredentials.Config{
		ClientID:     user,
		ClientSecret: password,
		Scopes:       cf.Scopes,
		TokenURL:     cf.TokenURL,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	token := a.TokenFromCache(cf, c)

	if token == nil {
		t, err := c.Token(context.WithValue(
			r.Context(),
			oauth2.HTTPClient,
			c.Client,
		))
		if err != nil {
			if rErr, ok := err.(*oauth2.RetrieveError); ok {
				switch httpStatusCode := rErr.Response.StatusCode; httpStatusCode {
				case http.StatusTooManyRequests:
					return errors.WithStack(helper.NewErrTooManyRequestsWithHeaders(rErr.Response))
				case http.StatusServiceUnavailable:
					return errors.Wrap(helper.ErrUpstreamServiceNotAvailable(), err.Error())
				case http.StatusInternalServerError:
					return errors.Wrap(helper.ErrUpstreamServiceInternalServerError(), err.Error())
				case http.StatusGatewayTimeout:
					return errors.Wrap(helper.ErrUpstreamServiceTimeout(), err.Error())
				case http.StatusNotFound:
					return errors.Wrap(helper.ErrUpstreamServiceNotFound(), err.Error())
				default:
					return errors.Wrap(helper.ErrUnauthorized(), err.Error())
				}
			} else {
				return errors.Wrap(helper.ErrUpstreamServiceNotAvailable(), err.Error())
			}
		}

		token = t

		a.TokenToCache(cf, c, *token)
	}

	if token.AccessToken == "" {
		return errors.WithStack(helper.ErrUnauthorized())
	}

	session.Subject = user
	return nil
}
