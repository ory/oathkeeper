// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package mutate

import (
	"bytes"
	"crypto/md5" //nolint:gosec
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/httpx"
)

const (
	ErrMalformedResponseFromUpstreamAPI = "The call to an external API returned an invalid JSON object"
	ErrMissingAPIURL                    = "Missing URL in mutator configuration"
	ErrInvalidAPIURL                    = "Invalid URL in mutator configuration"
	ErrNon200ResponseFromAPI            = "The call to an external API returned a non-200 HTTP response"
	ErrInvalidCredentials               = "Invalid credentials were provided in mutator configuration" //nolint:gosec // user-facing error text
	ErrNoCredentialsProvided            = "No credentials were provided in mutator configuration"
	contentTypeHeaderKey                = "Content-Type"
	contentTypeJSONHeaderValue          = "application/json"
)

type MutatorHydrator struct {
	c configuration.Provider
	d mutatorHydratorDependencies

	hydrateCache *ristretto.Cache[string, *authn.AuthenticationSession]
}

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type auth struct {
	Basic BasicAuth `json:"basic"`
}

type retryConfig struct {
	MaxDelay    string `json:"max_delay"`
	GiveUpAfter string `json:"give_up_after"`
}

type externalAPIConfig struct {
	URL   string       `json:"url"`
	Auth  *auth        `json:"auth"`
	Retry *retryConfig `json:"retry"`
}

type cacheConfig struct {
	Enabled bool   `json:"enabled"`
	TTL     string `json:"ttl"`

	ttl time.Duration
}

type MutatorHydratorConfig struct {
	Api   externalAPIConfig `json:"api"`
	Cache cacheConfig       `json:"cache"`
}

type mutatorHydratorDependencies interface {
	x.RegistryLogger
	Tracer() trace.Tracer
}

func NewMutatorHydrator(c configuration.Provider, d mutatorHydratorDependencies) *MutatorHydrator {
	cache, _ := ristretto.NewCache(&ristretto.Config[string, *authn.AuthenticationSession]{
		// This will hold about 1000 unique mutation responses.
		NumCounters: 10000,
		// Allocate a max of 32MB
		MaxCost: 1 << 25,
		// This is a best-practice value.
		BufferItems: 64,
	})
	return &MutatorHydrator{
		c:            c,
		d:            d,
		hydrateCache: cache,
	}
}

func (a *MutatorHydrator) GetID() string {
	return "hydrator"
}

func (a *MutatorHydrator) cacheKey(config *MutatorHydratorConfig, session string) string {
	return fmt.Sprintf("%s|%x", config.Api.URL, md5.Sum([]byte(session))) //nolint:gosec
}

func (a *MutatorHydrator) hydrateFromCache(config *MutatorHydratorConfig, session string) (*authn.AuthenticationSession, bool) {
	if !config.Cache.Enabled {
		return nil, false
	}

	item, found := a.hydrateCache.Get(a.cacheKey(config, session))
	if !found {
		return nil, false
	}

	return item.Copy(), true
}

func (a *MutatorHydrator) hydrateToCache(config *MutatorHydratorConfig, key string, session *authn.AuthenticationSession) {
	if !config.Cache.Enabled {
		return
	}

	if a.hydrateCache.SetWithTTL(a.cacheKey(config, key), session.Copy(), 0, config.Cache.ttl) {
		a.d.Logger().Debug("Cache reject item")
	}
}

func (a *MutatorHydrator) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	cfg, err := a.Config(config)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(session); err != nil {
		return errors.WithStack(err)
	}

	encodedSession := b.String()
	if cacheSession, ok := a.hydrateFromCache(cfg, encodedSession); ok {
		*session = *cacheSession
		return nil
	}

	if cfg.Api.URL == "" {
		return errors.New(ErrMissingAPIURL)
	} else if _, err := url.ParseRequestURI(cfg.Api.URL); err != nil {
		return errors.New(ErrInvalidAPIURL)
	}
	req, err := http.NewRequest("POST", cfg.Api.URL, &b)
	if err != nil {
		return errors.WithStack(err)
	}

	if r.URL != nil {
		q := r.URL.Query()
		req.URL.RawQuery = q.Encode()
	}

	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	if cfg.Api.Auth != nil {
		credentials := cfg.Api.Auth.Basic
		req.SetBasicAuth(credentials.Username, credentials.Password)
	}
	req.Header.Set(contentTypeHeaderKey, contentTypeJSONHeaderValue)

	client := http.DefaultClient
	if a.d.Tracer() != nil {
		client = otelhttp.DefaultClient
	}
	if cfg.Api.Retry != nil {
		giveUpAfter := time.Second
		maxRetryDelay := 100 * time.Millisecond
		if len(cfg.Api.Retry.MaxDelay) > 0 {
			if d, err := time.ParseDuration(cfg.Api.Retry.MaxDelay); err != nil {
				a.d.Logger().WithError(err).Warnf("Unable to parse max_delay in the Hydrator Mutator, falling back to default (%v).", maxRetryDelay)
			} else {
				maxRetryDelay = d
			}
		}
		if len(cfg.Api.Retry.GiveUpAfter) > 0 {
			if d, err := time.ParseDuration(cfg.Api.Retry.GiveUpAfter); err != nil {
				a.d.Logger().WithError(err).Warnf("Unable to parse give_up_after in the Hydrator Mutator, falling back to default (%v).", giveUpAfter)
			} else {
				giveUpAfter = d
			}
		}
		clientOpts := []httpx.ResilientOptions{
			httpx.ResilientClientWithConnectionTimeout(giveUpAfter),
			httpx.ResilientClientWithMaxRetryWait(maxRetryDelay)}
		client = httpx.NewResilientClient(clientOpts...).StandardClient()
	}

	res, err := client.Do(req.WithContext(r.Context()))
	if err != nil {
		return errors.WithStack(err)
	}
	defer res.Body.Close() //nolint:errcheck

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized:
		if cfg.Api.Auth != nil {
			return errors.New(ErrInvalidCredentials)
		} else {
			return errors.New(ErrNoCredentialsProvided)
		}
	default:
		return errors.New(ErrNon200ResponseFromAPI)
	}

	sessionFromUpstream := authn.AuthenticationSession{}
	err = json.NewDecoder(res.Body).Decode(&sessionFromUpstream)
	if err != nil {
		return errors.WithStack(err)
	}
	if sessionFromUpstream.Subject != session.Subject {
		return errors.New(ErrMalformedResponseFromUpstreamAPI)
	}
	*session = sessionFromUpstream

	a.hydrateToCache(cfg, encodedSession, session)

	return nil
}

func (a *MutatorHydrator) Validate(config json.RawMessage) error {
	if !a.c.MutatorIsEnabled(a.GetID()) {
		return NewErrMutatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *MutatorHydrator) Config(config json.RawMessage) (*MutatorHydratorConfig, error) {
	var c MutatorHydratorConfig
	if err := a.c.MutatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrMutatorMisconfigured(a, err)
	}

	if c.Cache.Enabled {
		var err error
		c.Cache.ttl, err = time.ParseDuration(c.Cache.TTL)
		if err != nil {
			a.d.Logger().WithError(err).WithField("ttl", c.Cache.TTL).Error("Unable to parse cache ttl in the Hydrator Mutator.")
			return nil, NewErrMutatorMisconfigured(a, err)
		}

		if c.Cache.ttl == 0 {
			c.Cache.ttl = time.Minute
		}
	}

	return &c, nil
}
