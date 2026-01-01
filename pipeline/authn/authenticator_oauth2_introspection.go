// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/fosite"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/x/header"
	"github.com/ory/x/httpx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/stringslice"
)

type AuthenticatorOAuth2IntrospectionConfiguration struct {
	Scopes                      []string                                              `json:"required_scope"`
	Audience                    []string                                              `json:"target_audience"`
	Issuers                     []string                                              `json:"trusted_issuers"`
	PreAuth                     *AuthenticatorOAuth2IntrospectionPreAuthConfiguration `json:"pre_authorization"`
	ScopeStrategy               string                                                `json:"scope_strategy"`
	IntrospectionURL            string                                                `json:"introspection_url"`
	PreserveHost                bool                                                  `json:"preserve_host"`
	BearerTokenLocation         *helper.BearerTokenLocation                           `json:"token_from"`
	Prefix                      string                                                `json:"prefix"`
	IntrospectionRequestHeaders map[string]string                                     `json:"introspection_request_headers"`
	Retry                       *AuthenticatorOAuth2IntrospectionRetryConfiguration   `json:"retry"`
	Cache                       cacheConfig                                           `json:"cache"`
}

type AuthenticatorOAuth2IntrospectionPreAuthConfiguration struct {
	Enabled      bool     `json:"enabled"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Audience     string   `json:"audience"`
	Scope        []string `json:"scope"`
	TokenURL     string   `json:"token_url"`
}

type AuthenticatorOAuth2IntrospectionRetryConfiguration struct {
	Timeout string `json:"max_delay"`
	MaxWait string `json:"give_up_after"`
}

type cacheConfig struct {
	Enabled bool   `json:"enabled"`
	TTL     string `json:"ttl"`
	MaxCost int    `json:"max_cost"`
}

type AuthenticatorOAuth2Introspection struct {
	c configuration.Provider

	clientMap map[string]*http.Client
	mu        sync.RWMutex

	tokenCache *ristretto.Cache[string, []byte]
	cacheTTL   *time.Duration
	logger     *logrusx.Logger
	provider   trace.TracerProvider
}

func NewAuthenticatorOAuth2Introspection(c configuration.Provider, l *logrusx.Logger, p trace.TracerProvider) *AuthenticatorOAuth2Introspection {
	return &AuthenticatorOAuth2Introspection{c: c, logger: l, provider: p, clientMap: make(map[string]*http.Client)}
}

func (a *AuthenticatorOAuth2Introspection) GetID() string {
	return "oauth2_introspection"
}

type Audience []string

type AuthenticatorOAuth2IntrospectionResult struct {
	Active    bool                   `json:"active"`
	Extra     map[string]interface{} `json:"ext"`
	Subject   string                 `json:"sub,omitempty"`
	Username  string                 `json:"username"`
	Audience  Audience               `json:"aud,omitempty"`
	TokenType string                 `json:"token_type"`
	Issuer    string                 `json:"iss"`
	ClientID  string                 `json:"client_id,omitempty"`
	Scope     string                 `json:"scope,omitempty"`
	Expires   int64                  `json:"exp"`
	TokenUse  string                 `json:"token_use"`
}

func (a *Audience) UnmarshalJSON(b []byte) error {
	var errUnsupportedType = errors.New("Unsupported aud type, only string or []string are allowed")

	var jsonObject interface{}
	err := json.Unmarshal(b, &jsonObject)
	if err != nil {
		return err
	}

	switch o := jsonObject.(type) {
	case string:
		*a = Audience{o}
		return nil
	case []interface{}:
		s := make(Audience, 0, len(o))
		for _, v := range o {
			value, ok := v.(string)
			if !ok {
				return errUnsupportedType
			}
			s = append(s, value)
		}
		*a = s
		return nil
	}

	return errUnsupportedType
}

func (a *AuthenticatorOAuth2Introspection) tokenFromCache(config *AuthenticatorOAuth2IntrospectionConfiguration, token string, ss fosite.ScopeStrategy) *AuthenticatorOAuth2IntrospectionResult {
	if !config.Cache.Enabled {
		return nil
	}

	if ss == nil && len(config.Scopes) > 0 {
		return nil
	}

	i, found := a.tokenCache.Get(token)
	if !found {
		return nil
	}

	var v AuthenticatorOAuth2IntrospectionResult
	if err := json.Unmarshal(i, &v); err != nil {
		return nil
	}
	return &v
}

func (a *AuthenticatorOAuth2Introspection) tokenToCache(config *AuthenticatorOAuth2IntrospectionConfiguration, i *AuthenticatorOAuth2IntrospectionResult, token string, ss fosite.ScopeStrategy) {
	if !config.Cache.Enabled {
		return
	}

	if ss == nil && len(config.Scopes) > 0 {
		return
	}

	if v, err := json.Marshal(i); err != nil {
		return
	} else if a.cacheTTL != nil {
		a.tokenCache.SetWithTTL(token, v, 1, *a.cacheTTL)
	} else {
		a.tokenCache.Set(token, v, 1)
	}
}

func (a *AuthenticatorOAuth2Introspection) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) (err error) {
	tp := trace.SpanFromContext(r.Context()).TracerProvider()
	ctx, span := tp.Tracer("oauthkeeper/pipeline/authn").Start(r.Context(), "pipeline.authn.AuthenticatorOAuth2Introspection.Authenticate")
	defer otelx.End(span, &err)
	r = r.WithContext(ctx)

	cf, client, err := a.Config(config)
	if err != nil {
		return err
	}

	token := helper.BearerTokenFromRequest(r, cf.BearerTokenLocation)
	if token == "" || !strings.HasPrefix(token, cf.Prefix) {
		return errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	ss := a.c.ToScopeStrategy(cf.ScopeStrategy, "authenticators.oauth2_introspection.config.scope_strategy")

	i := a.tokenFromCache(cf, token, ss)
	inCache := i != nil

	// If the token can not be found, and the scope strategy is nil, and the required scope list
	// is not empty, then we can not use the cache.
	if !inCache {
		body := url.Values{"token": {token}}
		if ss == nil {
			body.Add("scope", strings.Join(cf.Scopes, " "))
		}

		introspectReq, err := http.NewRequestWithContext(ctx, http.MethodPost, cf.IntrospectionURL, strings.NewReader(body.Encode()))
		if err != nil {
			return errors.WithStack(err)
		}

		for key, value := range cf.IntrospectionRequestHeaders {
			introspectReq.Header.Set(key, value)
		}
		// set/override the content-type header
		introspectReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		if cf.PreserveHost {
			introspectReq.Header.Set(header.XForwardedHost, r.Host)
		}

		resp, err := client.Do(introspectReq)
		if err != nil {
			return errors.WithStack(err)
		}
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode != http.StatusOK {
			return errors.Errorf("Introspection returned status code %d but expected %d", resp.StatusCode, http.StatusOK)
		}

		if err := json.NewDecoder(resp.Body).Decode(&i); err != nil {
			return errors.WithStack(err)
		}
	}

	if len(i.TokenUse) > 0 && i.TokenUse != "access_token" {
		return errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Use of introspected token is not an access token but \"%s\"", i.TokenUse)))
	}

	if !i.Active {
		return errors.WithStack(helper.ErrUnauthorized.WithReason("Access token is not active"))
	}

	if i.Expires > 0 && time.Unix(i.Expires, 0).Before(time.Now()) {
		return errors.WithStack(helper.ErrUnauthorized.WithReason("Access token expired"))
	}

	for _, audience := range cf.Audience {
		if !stringslice.Has(i.Audience, audience) {
			return errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Token audience is not intended for target audience %s", audience)))
		}
	}

	if len(cf.Issuers) > 0 {
		if !stringslice.Has(cf.Issuers, i.Issuer) {
			return errors.WithStack(helper.ErrForbidden.WithReason("Token issuer does not match any trusted issuer"))
		}
	}

	if ss != nil {
		for _, scope := range cf.Scopes {
			if !ss(strings.Split(i.Scope, " "), scope) {
				return errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Scope %s was not granted", scope)))
			}
		}
	}

	if !inCache {
		a.tokenToCache(cf, i, token, ss)
	}

	if len(i.Extra) == 0 {
		i.Extra = map[string]interface{}{}
	}

	i.Extra["username"] = i.Username
	i.Extra["client_id"] = i.ClientID
	i.Extra["scope"] = i.Scope

	if len(i.Audience) != 0 {
		i.Extra["aud"] = i.Audience
	}

	session.Subject = i.Subject
	session.Extra = i.Extra

	return nil
}

func (a *AuthenticatorOAuth2Introspection) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	_, _, err := a.Config(config)
	return err
}

func (a *AuthenticatorOAuth2Introspection) Config(config json.RawMessage) (*AuthenticatorOAuth2IntrospectionConfiguration, *http.Client, error) {
	var c AuthenticatorOAuth2IntrospectionConfiguration
	if err := a.c.AuthenticatorConfig(a.GetID(), config, &c); err != nil {
		return nil, nil, NewErrAuthenticatorMisconfigured(a, err)
	}

	rawKey, err := json.Marshal(&c)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	clientKey := fmt.Sprintf("%x", md5.Sum(rawKey)) //nolint:gosec
	a.mu.RLock()
	client, ok := a.clientMap[clientKey]
	a.mu.RUnlock()

	if !ok || client == nil {
		a.logger.Debug("Initializing http client")
		var rt http.RoundTripper
		if c.PreAuth != nil && c.PreAuth.Enabled {
			var ep url.Values

			if c.PreAuth.Audience != "" {
				ep = url.Values{"audience": {c.PreAuth.Audience}}
			}

			rt = (&clientcredentials.Config{
				ClientID:       c.PreAuth.ClientID,
				ClientSecret:   c.PreAuth.ClientSecret,
				Scopes:         c.PreAuth.Scope,
				EndpointParams: ep,
				TokenURL:       c.PreAuth.TokenURL,
			}).Client(context.Background()).Transport
		}

		if c.Retry == nil {
			c.Retry = &AuthenticatorOAuth2IntrospectionRetryConfiguration{Timeout: "500ms", MaxWait: "1s"}
		} else {
			if c.Retry.Timeout == "" {
				c.Retry.Timeout = "500ms"
			}
			if c.Retry.MaxWait == "" {
				c.Retry.MaxWait = "1s"
			}
		}
		duration, err := time.ParseDuration(c.Retry.Timeout)
		if err != nil {
			return nil, nil, errors.WithStack(err)
		}
		timeout := time.Millisecond * duration

		maxWait, err := time.ParseDuration(c.Retry.MaxWait)
		if err != nil {
			return nil, nil, errors.WithStack(err)
		}

		client = httpx.NewResilientClient(
			httpx.ResilientClientWithMaxRetryWait(maxWait),
			httpx.ResilientClientWithConnectionTimeout(timeout),
		).StandardClient()
		client.Transport = otelhttp.NewTransport(rt, otelhttp.WithTracerProvider(a.provider))
		a.mu.Lock()
		a.clientMap[clientKey] = client
		a.mu.Unlock()
	}

	if c.Cache.TTL != "" {
		cacheTTL, err := time.ParseDuration(c.Cache.TTL)
		if err != nil {
			return nil, nil, err
		}

		// clear cache if previous ttl was longer (or none)
		if a.tokenCache != nil {
			if a.cacheTTL == nil || (a.cacheTTL != nil && a.cacheTTL.Seconds() > cacheTTL.Seconds()) {
				a.tokenCache.Clear()
			}
		}

		a.cacheTTL = &cacheTTL
	}

	if a.tokenCache == nil {
		cost := int64(c.Cache.MaxCost)
		if cost == 0 {
			cost = 100000000
		}
		a.logger.Debugf("Creating cache with max cost: %d", c.Cache.MaxCost)
		cache, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
			// This will hold about 1000 unique mutation responses.
			NumCounters: cost * 10,
			// Allocate a max
			MaxCost: cost,
			// This is a best-practice value.
			BufferItems: 64,
			Cost: func(value []byte) int64 {
				return 1
			},
			IgnoreInternalCost: true,
		})
		if err != nil {
			return nil, nil, err
		}

		a.tokenCache = cache
	}

	return &c, client, nil
}
