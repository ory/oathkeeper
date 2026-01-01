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
	"golang.org/x/oauth2"

	"github.com/ory/x/logrusx"

	"github.com/ory/x/httpx"

	"github.com/ory/oathkeeper/driver/configuration"

	"github.com/ory/oathkeeper/pipeline"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/oathkeeper/helper"
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

type AuthenticatorOAuth2ClientCredentials struct {
	c      configuration.Provider
	client *http.Client

	tokenCache *ristretto.Cache[string, []byte]
	cacheTTL   *time.Duration
	logger     *logrusx.Logger
}

type AuthenticatorOAuth2ClientCredentialsRetryConfiguration struct {
	Timeout string `json:"max_delay"`
	MaxWait string `json:"give_up_after"`
}

func NewAuthenticatorOAuth2ClientCredentials(c configuration.Provider, logger *logrusx.Logger) *AuthenticatorOAuth2ClientCredentials {
	return &AuthenticatorOAuth2ClientCredentials{c: c, logger: logger}
}

func (a *AuthenticatorOAuth2ClientCredentials) GetID() string {
	return "oauth2_client_credentials"
}

func (a *AuthenticatorOAuth2ClientCredentials) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *AuthenticatorOAuth2ClientCredentials) Config(config json.RawMessage) (*AuthenticatorOAuth2Configuration, error) {
	const (
		defaultTimeout = "1s"
		defaultMaxWait = "2s"
	)
	var c AuthenticatorOAuth2Configuration
	if err := a.c.AuthenticatorConfig(a.GetID(), config, &c); err != nil {
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
	duration, err := time.ParseDuration(c.Retry.Timeout)
	if err != nil {
		return nil, err
	}

	maxWait, err := time.ParseDuration(c.Retry.MaxWait)
	if err != nil {
		return nil, err
	}
	timeout := time.Millisecond * duration
	a.client = httpx.NewResilientClient(
		httpx.ResilientClientWithMaxRetryWait(maxWait),
		httpx.ResilientClientWithConnectionTimeout(timeout),
	).StandardClient()

	if c.Cache.TTL != "" {
		cacheTTL, err := time.ParseDuration(c.Cache.TTL)
		if err != nil {
			return nil, err
		}
		a.cacheTTL = &cacheTTL
	}

	if a.tokenCache == nil {
		maxTokens := int64(c.Cache.MaxTokens)
		if maxTokens == 0 {
			maxTokens = 1000
		}
		a.logger.Debugf("Creating cache with max tokens: %d", maxTokens)
		cache, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
			// This will hold about 1000 unique mutation responses.
			NumCounters: 10 * maxTokens,
			// Allocate a maximum amount of tokens to cache
			MaxCost: maxTokens,
			// This is a best-practice value.
			BufferItems: 64,
			// Use a static cost of 1, so we can limit the amount of tokens that can be stored
			Cost: func(value []byte) int64 {
				return 1
			},
			IgnoreInternalCost: true,
		})
		if err != nil {
			return nil, err
		}

		a.tokenCache = cache
	}

	return &c, nil
}

func clientCredentialsConfigToKey(cc clientcredentials.Config) string {
	return fmt.Sprintf("%s|%s|%s:%s", cc.TokenURL, strings.Join(cc.Scopes, " "), cc.ClientID, cc.ClientSecret)
}

func (a *AuthenticatorOAuth2ClientCredentials) tokenFromCache(config *AuthenticatorOAuth2Configuration, clientCredentials clientcredentials.Config) *oauth2.Token {
	if !config.Cache.Enabled {
		return nil
	}

	i, found := a.tokenCache.Get(clientCredentialsConfigToKey(clientCredentials))
	if !found {
		return nil
	}

	var v oauth2.Token
	if err := json.Unmarshal(i, &v); err != nil {
		return nil
	}
	return &v
}

func (a *AuthenticatorOAuth2ClientCredentials) tokenToCache(config *AuthenticatorOAuth2Configuration, clientCredentials clientcredentials.Config, token oauth2.Token) {
	if !config.Cache.Enabled {
		return
	}

	key := clientCredentialsConfigToKey(clientCredentials)

	if v, err := json.Marshal(token); err != nil {
		return
	} else if a.cacheTTL != nil {
		// Allow up-to at most the cache TTL, otherwise use token expiry
		ttl := time.Until(token.Expiry)
		if ttl > *a.cacheTTL {
			ttl = *a.cacheTTL
		}

		a.tokenCache.SetWithTTL(key, v, 1, ttl)
	} else {
		// If token has no expiry apply the same to the cache
		ttl := time.Duration(0)
		if !token.Expiry.IsZero() {
			ttl = time.Until(token.Expiry)
		}

		a.tokenCache.SetWithTTL(key, v, 1, ttl)
	}
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
		return errors.Wrap(helper.ErrUnauthorized, err.Error())
	}

	password, err = url.QueryUnescape(password)
	if err != nil {
		return errors.Wrap(helper.ErrUnauthorized, err.Error())
	}

	c := clientcredentials.Config{
		ClientID:     user,
		ClientSecret: password,
		Scopes:       cf.Scopes,
		TokenURL:     cf.TokenURL,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}

	token := a.tokenFromCache(cf, c)

	if token == nil {
		t, err := c.Token(context.WithValue(
			r.Context(),
			oauth2.HTTPClient,
			c.Client,
		))

		if err != nil {
			if rErr, ok := err.(*oauth2.RetrieveError); ok {
				switch httpStatusCode := rErr.Response.StatusCode; httpStatusCode {
				case http.StatusServiceUnavailable:
					return errors.Wrap(helper.ErrUpstreamServiceNotAvailable, err.Error())
				case http.StatusInternalServerError:
					return errors.Wrap(helper.ErrUpstreamServiceInternalServerError, err.Error())
				case http.StatusGatewayTimeout:
					return errors.Wrap(helper.ErrUpstreamServiceTimeout, err.Error())
				case http.StatusNotFound:
					return errors.Wrap(helper.ErrUpstreamServiceNotFound, err.Error())
				default:
					return errors.Wrap(helper.ErrUnauthorized, err.Error())
				}
			} else {
				return errors.Wrap(helper.ErrUpstreamServiceNotAvailable, err.Error())
			}
		}

		token = t

		a.tokenToCache(cf, c, *token)
	}

	if token.AccessToken == "" {
		return errors.WithStack(helper.ErrUnauthorized)
	}

	session.Subject = user
	return nil
}
