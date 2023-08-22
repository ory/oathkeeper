// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/ory/oathkeeper/driver/configuration"

	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/x"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/rule"
)

type proxyRegistry interface {
	x.RegistryLogger
	x.RegistryWriter
	ProxyRequestHandler() RequestHandler
	RuleMatcher() rule.Matcher
}

func NewProxy(r proxyRegistry, c configuration.Provider) *Proxy {
	return &Proxy{r: r, c: c}
}

type Proxy struct {
	r proxyRegistry
	c configuration.Provider
}

type key int

const (
	director key = iota + 1
	ContextKeyMatchedRule
	ContextKeySession
)

func (d *Proxy) RoundTrip(r *http.Request) (*http.Response, error) {
	rw := NewSimpleResponseWriter()
	fields := map[string]interface{}{
		"http_method":     r.Method,
		"http_url":        r.URL.String(),
		"http_host":       r.Host,
		"http_user_agent": r.UserAgent(),
	}

	if sess, ok := r.Context().Value(ContextKeySession).(*authn.AuthenticationSession); ok {
		fields["subject"] = sess.Subject
	}

	rl, _ := r.Context().Value(ContextKeyMatchedRule).(*rule.Rule)

	if err, ok := r.Context().Value(director).(error); ok && err != nil {
		d.r.Logger().WithError(err).
			WithFields(fields).
			WithField("granted", false).
			Warn("Access request denied")

		d.r.ProxyRequestHandler().HandleError(rw, r, rl, err)

		return &http.Response{
			StatusCode: rw.code,
			Body:       io.NopCloser(rw.buffer),
			Header:     rw.header,
		}, nil
	} else if err == nil {
		res, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			d.r.Logger().
				WithError(errors.WithStack(err)).
				WithField("granted", false).
				WithFields(fields).
				Warn("Access request denied because roundtrip failed")
			// don't need to return because covered in next line
		} else {
			d.r.Logger().
				WithField("granted", true).
				WithFields(fields).
				Info("Access request granted")
		}

		return res, err
	}

	err := errors.New("Unable to type assert context")
	d.r.Logger().
		WithError(err).
		WithField("granted", false).
		WithFields(fields).
		Warn("Unable to type assert context")

	d.r.ProxyRequestHandler().HandleError(rw, r, rl, err)

	return &http.Response{
		StatusCode: rw.code,
		Body:       io.NopCloser(rw.buffer),
		Header:     rw.header,
	}, nil
}

func (d *Proxy) Rewrite(r *httputil.ProxyRequest) {
	if d.c.ProxyTrustForwardedHeaders() {
		headers := []string{
			"X-Forwarded-Host",
			"X-Forwarded-Proto",
			"X-Forwarded-For",
		}
		for _, h := range headers {
			if v := r.In.Header.Get(h); v != "" {
				r.Out.Header.Set(h, v)
			}
		}
	}

	EnrichRequestedURL(r)
	rl, err := d.r.RuleMatcher().Match(r.Out.Context(), r.Out.Method, r.Out.URL, rule.ProtocolHTTP)
	if err != nil {
		*r.Out = *r.Out.WithContext(context.WithValue(r.Out.Context(), director, err))
		return
	}

	*r.Out = *r.Out.WithContext(context.WithValue(r.Out.Context(), ContextKeyMatchedRule, rl))
	s, err := d.r.ProxyRequestHandler().HandleRequest(r.Out, rl)
	if err != nil {
		*r.Out = *r.Out.WithContext(context.WithValue(r.Out.Context(), director, err))
		return
	}
	*r.Out = *r.Out.WithContext(context.WithValue(r.Out.Context(), ContextKeySession, s))

	CopyHeaders(s.Header, r.Out)

	if err := ConfigureBackendURL(r.Out, rl); err != nil {
		*r.Out = *r.Out.WithContext(context.WithValue(r.Out.Context(), director, err))
		return
	}

	var en error // need to set it to error but with nil value
	*r.Out = *r.Out.WithContext(context.WithValue(r.Out.Context(), director, en))
}

func CopyHeaders(headers http.Header, r *http.Request) {
	if r.Header == nil {
		r.Header = make(map[string][]string)
	}
	for k, v := range headers {
		var val string
		if len(v) == 0 {
			val = ""
		} else {
			val = v[0]
		}
		r.Header.Set(k, val)
	}
}

// EnrichRequestedURL sets Scheme and Host values in a URL passed down by a http server. Per default, the URL
// does not contain host nor scheme values.
func EnrichRequestedURL(r *httputil.ProxyRequest) {
	r.Out.URL.Scheme = "http"
	r.Out.URL.Host = r.In.Host
	if r.In.TLS != nil || strings.EqualFold(r.In.Header.Get("X-Forwarded-Proto"), "https") {
		r.Out.URL.Scheme = "https"
	}
}

func ConfigureBackendURL(r *http.Request, rl *rule.Rule) error {
	if rl.Upstream.URL == "" {
		return errors.Errorf("Unable to forward the request because matched rule does not define an upstream URL")
	}

	p, err := url.Parse(rl.Upstream.URL)
	if err != nil {
		return errors.WithStack(err)
	}

	proxyHost := r.Host
	proxyPath := r.URL.Path

	backendHost := p.Host
	backendPath := p.Path
	backendScheme := p.Scheme

	forwardURL := r.URL
	forwardURL.Scheme = backendScheme
	forwardURL.Host = backendHost
	forwardURL.Path = "/" + strings.TrimLeft("/"+strings.Trim(backendPath, "/")+"/"+strings.TrimLeft(proxyPath, "/"), "/")

	if rl.Upstream.StripPath != "" {
		forwardURL.Path = strings.Replace(forwardURL.Path, "/"+strings.Trim(rl.Upstream.StripPath, "/"), "", 1)
	}

	r.Host = backendHost
	if rl.Upstream.PreserveHost {
		r.Host = proxyHost
	}

	return nil
}
