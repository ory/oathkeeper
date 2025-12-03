// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"crypto/md5" //#nosec G501 -- MD5 is used for cache key generation, not cryptography
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"

	"github.com/ory/oathkeeper/x/header"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/stringsx"

	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
)

func init() {
	gjson.AddModifier("this", func(json, arg string) string {
		return json
	})
}

type AuthenticatorCookieSessionFilter struct {
}

type AuthenticatorCookieSessionConfiguration struct {
	Only               []string                 `json:"only"`
	CheckSessionURL    string                   `json:"check_session_url"`
	PreserveQuery      bool                     `json:"preserve_query"`
	PreservePath       bool                     `json:"preserve_path"`
	ExtraFrom          string                   `json:"extra_from"`
	SubjectFrom        string                   `json:"subject_from"`
	PreserveHost       bool                     `json:"preserve_host"`
	ForwardHTTPHeaders []string                 `json:"forward_http_headers"`
	SetHeaders         map[string]string        `json:"additional_headers"`
	ForceMethod        string                   `json:"force_method"`
	Cache              cookieSessionCacheConfig `json:"cache"`
}

type cookieSessionCacheConfig struct {
	Enabled bool   `json:"enabled"`
	TTL     string `json:"ttl"`
	MaxCost int64  `json:"max_cost"`
}

func (a *AuthenticatorCookieSessionConfiguration) GetCheckSessionURL() string {
	return a.CheckSessionURL
}

func (a *AuthenticatorCookieSessionConfiguration) GetPreserveQuery() bool {
	return a.PreserveQuery
}

func (a *AuthenticatorCookieSessionConfiguration) GetPreservePath() bool {
	return a.PreservePath
}

func (a *AuthenticatorCookieSessionConfiguration) GetPreserveHost() bool {
	return a.PreserveHost
}

func (a *AuthenticatorCookieSessionConfiguration) GetForwardHTTPHeaders() []string {
	return a.ForwardHTTPHeaders
}

func (a *AuthenticatorCookieSessionConfiguration) GetSetHeaders() map[string]string {
	return a.SetHeaders
}

func (a *AuthenticatorCookieSessionConfiguration) GetForceMethod() string {
	return a.ForceMethod
}

type AuthenticatorCookieSession struct {
	c            configuration.Provider
	client       *http.Client
	tracer       trace.Tracer
	sessionCache *ristretto.Cache[string, []byte]
	cacheTTL     *time.Duration
	logger       *logrusx.Logger
}

var _ AuthenticatorForwardConfig = new(AuthenticatorCookieSessionConfiguration)

func NewAuthenticatorCookieSession(c configuration.Provider, logger *logrusx.Logger, provider trace.TracerProvider) *AuthenticatorCookieSession {
	return &AuthenticatorCookieSession{
		c:      c,
		logger: logger,
		client: &http.Client{
			Transport: otelhttp.NewTransport(
				http.DefaultTransport,
				otelhttp.WithTracerProvider(provider),
			),
		},
		tracer: provider.Tracer("oauthkeeper/pipeline/authn"),
	}
}

func (a *AuthenticatorCookieSession) GetID() string {
	return "cookie_session"
}

func (a *AuthenticatorCookieSession) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *AuthenticatorCookieSession) Config(config json.RawMessage) (*AuthenticatorCookieSessionConfiguration, error) {
	var c AuthenticatorCookieSessionConfiguration
	if err := a.c.AuthenticatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthenticatorMisconfigured(a, err)
	}

	if len(c.ExtraFrom) == 0 {
		c.ExtraFrom = "extra"
	}

	if len(c.SubjectFrom) == 0 {
		c.SubjectFrom = "subject"
	}

	// Add Authorization and Cookie headers for backward compatibility
	c.ForwardHTTPHeaders = append(c.ForwardHTTPHeaders, []string{header.Cookie}...)

	if c.Cache.TTL != "" {
		cacheTTL, err := time.ParseDuration(c.Cache.TTL)
		if err != nil {
			return nil, err
		}

		if a.sessionCache != nil {
			if a.cacheTTL == nil || (a.cacheTTL != nil && a.cacheTTL.Seconds() > cacheTTL.Seconds()) {
				a.sessionCache.Clear()
			}
		}

		a.cacheTTL = &cacheTTL
	}

	if a.sessionCache == nil {
		cost := c.Cache.MaxCost
		if cost == 0 {
			cost = 10000000
		}
		a.logger.Debugf("Creating session cache with max cost: %d", cost)
		cache, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
			NumCounters: cost * 10,
			MaxCost:     cost,
			BufferItems: 64,
			Cost: func(value []byte) int64 {
				return 1
			},
			IgnoreInternalCost: true,
		})
		if err != nil {
			return nil, err
		}

		a.sessionCache = cache
	}

	return &c, nil
}

func cookiesToCacheKey(cookies []*http.Cookie) string {
	var parts []string
	for _, cookie := range cookies {
		parts = append(parts, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(strings.Join(parts, "|")))) //#nosec G401 -- MD5 is used for cache key generation, not cryptography
}

type cachedSessionData struct {
	Subject string                 `json:"subject"`
	Extra   map[string]interface{} `json:"extra"`
}

func (a *AuthenticatorCookieSession) sessionFromCache(config *AuthenticatorCookieSessionConfiguration, r *http.Request) *cachedSessionData {
	if !config.Cache.Enabled {
		return nil
	}

	key := cookiesToCacheKey(r.Cookies())
	i, found := a.sessionCache.Get(key)
	if !found {
		return nil
	}

	var v cachedSessionData
	if err := json.Unmarshal(i, &v); err != nil {
		return nil
	}
	return &v
}

func (a *AuthenticatorCookieSession) sessionToCache(config *AuthenticatorCookieSessionConfiguration, r *http.Request, subject string, extra map[string]interface{}) {
	if !config.Cache.Enabled {
		return
	}

	key := cookiesToCacheKey(r.Cookies())
	data := cachedSessionData{
		Subject: subject,
		Extra:   extra,
	}

	if v, err := json.Marshal(data); err != nil {
		return
	} else if a.cacheTTL != nil {
		a.sessionCache.SetWithTTL(key, v, 1, *a.cacheTTL)
	} else {
		a.sessionCache.Set(key, v, 1)
	}
}

func (a *AuthenticatorCookieSession) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) (err error) {
	ctx, span := a.tracer.Start(r.Context(), "pipeline.authn.AuthenticatorCookieSession.Authenticate")
	defer otelx.End(span, &err)
	r = r.WithContext(ctx)

	cf, err := a.Config(config)
	if err != nil {
		return err
	}

	if !cookieSessionResponsible(r, cf.Only) {
		return errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	cachedSession := a.sessionFromCache(cf, r)
	if cachedSession != nil {
		session.Subject = cachedSession.Subject
		session.Extra = cachedSession.Extra
		return nil
	}

	body, err := forwardRequestToSessionStore(a.client, r, cf)
	if err != nil {
		return err
	}

	var (
		subject string
		extra   map[string]interface{}

		subjectRaw = []byte(stringsx.Coalesce(gjson.GetBytes(body, cf.SubjectFrom).Raw, "null"))
		extraRaw   = []byte(stringsx.Coalesce(gjson.GetBytes(body, cf.ExtraFrom).Raw, "null"))
	)

	if err = json.Unmarshal(subjectRaw, &subject); err != nil {
		return helper.ErrForbidden.WithReasonf("The configured subject_from GJSON path returned an error on JSON output: %s", err.Error()).WithDebugf("GJSON path: %s\nBody: %s\nResult: %s", cf.SubjectFrom, body, subjectRaw).WithTrace(err)
	}

	if err = json.Unmarshal(extraRaw, &extra); err != nil {
		return helper.ErrForbidden.WithReasonf("The configured extra_from GJSON path returned an error on JSON output: %s", err.Error()).WithDebugf("GJSON path: %s\nBody: %s\nResult: %s", cf.ExtraFrom, body, extraRaw).WithTrace(err)
	}

	session.Subject = subject
	session.Extra = extra

	a.sessionToCache(cf, r, subject, extra)

	return nil
}

func cookieSessionResponsible(r *http.Request, only []string) bool {
	if len(only) == 0 && len(r.Cookies()) > 0 {
		return true
	}

	for _, cookieName := range only {
		if _, err := r.Cookie(cookieName); err == nil {
			return true
		}
	}

	return false
}

func forwardRequestToSessionStore(client *http.Client, r *http.Request, cf AuthenticatorForwardConfig) (json.RawMessage, error) {
	req, err := PrepareRequest(r, cf)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req.WithContext(r.Context()))

	if err != nil {
		return nil, helper.ErrForbidden.WithReason(err.Error()).WithTrace(err)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return json.RawMessage{}, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to fetch cookie session context from remote: %+v", err))
		}
		return body, nil
	}

	switch res.StatusCode {
	case http.StatusTooManyRequests:
		return json.RawMessage{}, errors.WithStack(helper.ErrTooManyRequests.WithReason("Session store rate limit exceeded"))
	case http.StatusServiceUnavailable:
		return json.RawMessage{}, errors.WithStack(helper.ErrUpstreamServiceNotAvailable.WithReason("Session store is unavailable"))
	case http.StatusInternalServerError:
		return json.RawMessage{}, errors.WithStack(helper.ErrUpstreamServiceInternalServerError.WithReason("Session store returned internal server error"))
	case http.StatusGatewayTimeout:
		return json.RawMessage{}, errors.WithStack(helper.ErrUpstreamServiceTimeout.WithReason("Session store request timed out"))
	case http.StatusNotFound:
		return json.RawMessage{}, errors.WithStack(helper.ErrUpstreamServiceNotFound.WithReason("Session store endpoint not found"))
	default:
		return json.RawMessage{}, errors.WithStack(helper.ErrUnauthorized)
	}
}

func PrepareRequest(r *http.Request, cf AuthenticatorForwardConfig) (http.Request, error) {
	reqURL, err := url.Parse(cf.GetCheckSessionURL())
	if err != nil {
		return http.Request{}, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Unable to parse session check URL: %s", err))
	}

	if !cf.GetPreservePath() {
		reqURL.Path = r.URL.Path
	}

	if !cf.GetPreserveQuery() {
		reqURL.RawQuery = r.URL.RawQuery
	}

	m := cf.GetForceMethod()
	if m == "" {
		m = r.Method
	}

	req := http.Request{
		Method: m,
		URL:    reqURL,
		Header: http.Header{},
	}

	// We need to copy only essential and configurable headers
	for requested, v := range r.Header {
		for _, allowed := range cf.GetForwardHTTPHeaders() {
			// Check against canonical names of header
			if requested == header.Canonical(allowed) {
				req.Header[requested] = v
			}
		}
	}

	for k, v := range cf.GetSetHeaders() {
		req.Header.Set(k, v)
	}

	if cf.GetPreserveHost() {
		req.Header.Set(header.XForwardedHost, r.Host)
	}
	return req, nil
}
