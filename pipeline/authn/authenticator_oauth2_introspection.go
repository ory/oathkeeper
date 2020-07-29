package authn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/ory/go-convenience/stringslice"
	"github.com/ory/x/httpx"

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
}

type AuthenticatorOAuth2Introspection struct {
	c configuration.Provider

	client *http.Client

	tokenCache *ristretto.Cache
	cacheTTL   *time.Duration
}

func NewAuthenticatorOAuth2Introspection(c configuration.Provider) *AuthenticatorOAuth2Introspection {
	var rt http.RoundTripper
	cache, _ := ristretto.NewCache(&ristretto.Config{
		// This will hold about 1000 unique mutation responses.
		NumCounters: 10000,
		// Allocate a max of 32MB
		MaxCost: 1 << 25,
		// This is a best-practice value.
		BufferItems: 64,
	})
	return &AuthenticatorOAuth2Introspection{c: c, client: httpx.NewResilientClientLatencyToleranceSmall(rt), tokenCache: cache}
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
}

func (a *AuthenticatorOAuth2Introspection) tokenFromCache(config *AuthenticatorOAuth2IntrospectionConfiguration, token string) (*AuthenticatorOAuth2IntrospectionResult, bool) {
	if !config.Cache.Enabled {
		return nil, false
	}

	item, found := a.tokenCache.Get(token)
	if !found {
		return nil, false
	}

	i := item.(*AuthenticatorOAuth2IntrospectionResult)
	expires := time.Unix(i.Expires, 0)
	if expires.Before(time.Now()) {
		a.tokenCache.Del(token)
		return nil, false
	}

	return i, true
}

func (a *AuthenticatorOAuth2Introspection) tokenToCache(config *AuthenticatorOAuth2IntrospectionConfiguration, i *AuthenticatorOAuth2IntrospectionResult, token string) {
	if !config.Cache.Enabled {
		return
	}

	if a.cacheTTL != nil {
		a.tokenCache.SetWithTTL(token, i, 0, *a.cacheTTL)
	} else {
		a.tokenCache.Set(token, i, 0)
	}
}

func (a *AuthenticatorOAuth2Introspection) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	cf, err := a.Config(config)
	if err != nil {
		return err
	}

	token := helper.BearerTokenFromRequest(r, cf.BearerTokenLocation)
	if token == "" {
		return errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	ss := a.c.ToScopeStrategy(cf.ScopeStrategy, "authenticators.oauth2_introspection.scope_strategy")

	i, ok := a.tokenFromCache(cf, token)
	if !ok {
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
		resp, err := a.client.Do(introspectReq)
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

		if len(i.TokenType) > 0 && i.TokenType != "access_token" {
			return errors.WithStack(helper.ErrForbidden.WithReason(fmt.Sprintf("Introspected token is not an access token but \"%s\"", i.TokenType)))
		}

		if !i.Active {
			return errors.WithStack(helper.ErrUnauthorized.WithReason("Access token i says token is not active"))
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

		if len(i.Extra) == 0 {
			i.Extra = map[string]interface{}{}
		}

		i.Extra["username"] = i.Username
		i.Extra["client_id"] = i.ClientID
		i.Extra["scope"] = i.Scope

		a.tokenToCache(cf, i, token)
	}

	session.Subject = i.Subject
	session.Extra = i.Extra

	return nil
}

func (a *AuthenticatorOAuth2Introspection) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *AuthenticatorOAuth2Introspection) Config(config json.RawMessage) (*AuthenticatorOAuth2IntrospectionConfiguration, error) {
	var c AuthenticatorOAuth2IntrospectionConfiguration
	if err := a.c.AuthenticatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthenticatorMisconfigured(a, err)
	}

	var rt http.RoundTripper

	if c.PreAuth != nil && c.PreAuth.Enabled {
		rt = (&clientcredentials.Config{
			ClientID:     c.PreAuth.ClientID,
			ClientSecret: c.PreAuth.ClientSecret,
			Scopes:       c.PreAuth.Scope,
			TokenURL:     c.PreAuth.TokenURL,
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
		return nil, err
	}
	timeout := time.Millisecond * duration

	maxWait, err := time.ParseDuration(c.Retry.MaxWait)
	if err != nil {
		return nil, err
	}

	a.client = httpx.NewResilientClientLatencyToleranceConfigurable(rt, timeout, maxWait)

	if c.Cache.TTL != "" {
		cacheTTL, err := time.ParseDuration(c.Cache.TTL)
		if err != nil {
			return nil, err
		}
		a.cacheTTL = &cacheTTL
	}

	return &c, nil
}
