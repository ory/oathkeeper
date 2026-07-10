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
	"sync"
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
// Stateless-after-init design:
// - TokenCache is created once (lazily) the first time Config() is called,
//   based on c.Cache.max_tokens. After initialization, the struct remains
//   immutable and safe for concurrent use.
// - Token TTL is derived per call from the token expiry and the dynamic
//   config TTL; no per-instance mutable TTL state is stored.
//
type AuthenticatorOAuth2ClientCredentials struct {
	d          dependencies
	TokenCache *ristretto.Cache[string, []byte]
	mu         sync.Mutex // guards one-time TokenCache init
}

type AuthenticatorOAuth2ClientCredentialsRetryConfiguration struct {
	Timeout string `json:"max_delay"`
	MaxWait string `json:"give_up_after"`
}

// NewAuthenticatorOAuth2ClientCredentials returns an instance; cache is
// initialized on first Config() to respect user-defined max_tokens.
func NewAuthenticatorOAuth2ClientCredentials(d dependencies) *AuthenticatorOAuth2ClientCredentials {
	return &AuthenticatorOAuth2ClientCredentials{d: d}
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
// configuration. It also performs a one-time TokenCache initialization using
// c.Cache.max_tokens (default 1000) in a concurrency-safe manner.
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

	// One-time cache initialization honoring configured max_tokens
	if a.TokenCache == nil {
		a.mu.Lock()
		if a.TokenCache == nil {
			maxTokens := int64(c.Cache.MaxTokens)
			if maxTokens == 0 {
				maxTokens = 1000
			}
			cache, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
				// Frequency sketch size: 10x of MaxCost (ristretto best practice)
				NumCounters:        10 * maxTokens,
				MaxCost:            maxTokens,
				BufferItems:        64,
				Cost:               func(_ []byte) int64 { return 1 },
				IgnoreInternalCost: true,
			})
			if err != nil {
				a.mu.Unlock()
				return nil, err
			}
			a.TokenCache = cache
		}
		a.mu.Unlock()
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

	// TokenCache is immutable after initialization: no lock required.
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

	// Derive effective TTL from token expiry with a cap from config.Cache.TTL.
	// If the token has a zero expiry, we fall back to the configured TTL to
	// avoid caching with an unintended zero/forever TTL.
	if config.Cache.TTL != "" {
		if cacheTTL, parseErr := time.ParseDuration(config.Cache.TTL); parseErr == nil {
			var ttl time.Duration
			if token.Expiry.IsZero() {
				// Zero-expiry token: fall back to configured TTL
				ttl = cacheTTL
			} else {
				ttl = time.Until(token.Expiry)
				if ttl > cacheTTL {
					ttl = cacheTTL
				}
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
