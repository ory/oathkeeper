package authn

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ory/fosite"

	"github.com/dgraph-io/ristretto"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/go-convenience/stringslice"
	"github.com/ory/x/httpx"
	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
)

type AuthenticatorOAuth2IntrospectionConfiguration struct {
	Scopes                      []string                                              `json:"required_scope"`
	Audience                    []string                                              `json:"target_audience"`
	Issuers                     []string                                              `json:"trusted_issuers"`
	PreAuth                     *AuthenticatorOAuth2IntrospectionPreAuthConfiguration `json:"pre_authorization"`
	ScopeStrategy               string                                                `json:"scope_strategy"`
	IntrospectionURL            string                                                `json:"introspection_url"`
	BearerTokenLocation         *helper.BearerTokenLocation                           `json:"token_from"`
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

	tokenCache *ristretto.Cache
	cacheTTL   *time.Duration
	logger     *logrusx.Logger
}

func NewAuthenticatorOAuth2Introspection(c configuration.Provider, logger *logrusx.Logger) *AuthenticatorOAuth2Introspection {
	return &AuthenticatorOAuth2Introspection{c: c, logger: logger, clientMap: make(map[string]*http.Client)}
}

func (a *AuthenticatorOAuth2Introspection) GetID() string {
	return "oauth2_introspection"
}

type AuthenticatorOAuth2IntrospectionResult struct {
	Active    bool                   `json:"active"`
	Extra     map[string]interface{} `json:"ext"`
	Subject   string                 `json:"sub,omitempty"`
	Username  string                 `json:"username"`
	Audience  []string               `json:"aud"`
	TokenType string                 `json:"token_type"`
	Issuer    string                 `json:"iss"`
	ClientID  string                 `json:"client_id,omitempty"`
	Scope     string                 `json:"scope,omitempty"`
	Expires   int64                  `json:"exp"`
	TokenUse  string                 `json:"token_use"`
}

func (a *AuthenticatorOAuth2Introspection) tokenFromCache(config *AuthenticatorOAuth2IntrospectionConfiguration, token string, ss fosite.ScopeStrategy) *AuthenticatorOAuth2IntrospectionResult {
	if !config.Cache.Enabled {
		return nil
	}

	if ss == nil && len(config.Scopes) > 0 {
		return nil
	}

	item, found := a.tokenCache.Get(token)
	if !found {
		return nil
	}

	i, ok := item.(*AuthenticatorOAuth2IntrospectionResult)
	if !ok {
		return nil
	}

	return i
}

func (a *AuthenticatorOAuth2Introspection) tokenToCache(config *AuthenticatorOAuth2IntrospectionConfiguration, i *AuthenticatorOAuth2IntrospectionResult, token string, ss fosite.ScopeStrategy) {
	if !config.Cache.Enabled {
		return
	}

	if ss == nil && len(config.Scopes) > 0 {
		return
	}

	if a.cacheTTL != nil {
		a.tokenCache.SetWithTTL(token, i, 1, *a.cacheTTL)
	} else {
		a.tokenCache.Set(token, i, 1)
	}
}

func (a *AuthenticatorOAuth2Introspection) traceRequest(ctx context.Context, req *http.Request) func() {
	tracer := opentracing.GlobalTracer()
	if tracer == nil {
		return func() {}
	}

	parentSpan := opentracing.SpanFromContext(ctx)
	opts := make([]opentracing.StartSpanOption, 0, 1)
	if parentSpan != nil {
		opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
	}

	urlStr := req.URL.String()
	clientSpan := tracer.StartSpan(req.Method+" "+urlStr, opts...)

	ext.SpanKindRPCClient.Set(clientSpan)
	ext.HTTPUrl.Set(clientSpan, urlStr)
	ext.HTTPMethod.Set(clientSpan, req.Method)

	_ = tracer.Inject(clientSpan.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	return clientSpan.Finish
}

func (a *AuthenticatorOAuth2Introspection) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	cf, client, err := a.Config(config)
	if err != nil {
		return err
	}

	token := helper.BearerTokenFromRequest(r, cf.BearerTokenLocation)
	if token == "" {
		return errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	ss := a.c.ToScopeStrategy(cf.ScopeStrategy, "authenticators.oauth2_introspection.config.scope_strategy")

	i := a.tokenFromCache(cf, token, ss)

	// If the token can not be found, and the scope strategy is nil, and the required scope list
	// is not empty, then we can not use the cache.
	if i == nil {
		body := url.Values{"token": {token}}
		if ss == nil {
			body.Add("scope", strings.Join(cf.Scopes, " "))
		}

		introspectReq, err := http.NewRequest(http.MethodPost, cf.IntrospectionURL, strings.NewReader(body.Encode()))
		if err != nil {
			return errors.WithStack(err)
		}

		for key, value := range cf.IntrospectionRequestHeaders {
			introspectReq.Header.Set(key, value)
		}
		// set/override the content-type header
		introspectReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// add tracing
		closeSpan := a.traceRequest(r.Context(), introspectReq)

		resp, err := client.Do(introspectReq.WithContext(r.Context()))

		// close the span so it represents just the http request
		closeSpan()
		if err != nil {
			return errors.WithStack(err)
		}
		defer resp.Body.Close()

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
			return errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Token issuer does not match any trusted issuer")))
		}
	}

	if ss != nil {
		for _, scope := range cf.Scopes {
			if !ss(strings.Split(i.Scope, " "), scope) {
				return errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Scope %s was not granted", scope)))
			}
		}
	}

	a.tokenToCache(cf, i, token, ss)

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

	clientKey := fmt.Sprintf("%x", md5.Sum(rawKey))
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

		client = httpx.NewResilientClientLatencyToleranceConfigurable(rt, timeout, maxWait)
		a.mu.Lock()
		a.clientMap[clientKey] = client
		a.mu.Unlock()
	}

	if c.Cache.TTL != "" {
		cacheTTL, err := time.ParseDuration(c.Cache.TTL)
		if err != nil {
			return nil, nil, err
		}
		a.cacheTTL = &cacheTTL
	}

	if a.tokenCache == nil {
		cost := int64(c.Cache.MaxCost)
		if cost == 0 {
			cost = 100000000
		}
		a.logger.Debugf("Creating cache with max cost: %d", c.Cache.MaxCost)
		cache, err := ristretto.NewCache(&ristretto.Config{
			// This will hold about 1000 unique mutation responses.
			NumCounters: 10000,
			// Allocate a max
			MaxCost: cost,
			// This is a best-practice value.
			BufferItems: 64,
		})
		if err != nil {
			return nil, nil, err
		}

		a.tokenCache = cache
	}

	return &c, client, nil
}
