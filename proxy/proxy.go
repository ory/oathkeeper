/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package proxy

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

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

func NewProxy(r proxyRegistry) *Proxy {
	return &Proxy{r: r}
}

type Proxy struct {
	r proxyRegistry
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
			Body:       ioutil.NopCloser(rw.buffer),
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
				Warn("Access request granted")
		}

		return res, err
	}

	err := errors.New("Unable to type assert context")
	d.r.Logger().
		WithError(err).
		WithField("granted", false).
		WithFields(fields).
		Warn("Unable to type assert context")

	// add tracing
	closeSpan := x.TraceRequest(r.Context(), r)
	defer closeSpan()

	d.r.ProxyRequestHandler().HandleError(rw, r, rl, err)

	return &http.Response{
		StatusCode: rw.code,
		Body:       ioutil.NopCloser(rw.buffer),
		Header:     rw.header,
	}, nil
}

func (d *Proxy) Director(r *http.Request) {
	EnrichRequestedURL(r)
	rl, err := d.r.RuleMatcher().Match(r.Context(), r.Method, r.URL)
	if err != nil {
		*r = *r.WithContext(context.WithValue(r.Context(), director, err))
		return
	}

	*r = *r.WithContext(context.WithValue(r.Context(), ContextKeyMatchedRule, rl))
	s, err := d.r.ProxyRequestHandler().HandleRequest(r, rl)
	if err != nil {
		*r = *r.WithContext(context.WithValue(r.Context(), director, err))
		return
	}
	*r = *r.WithContext(context.WithValue(r.Context(), ContextKeySession, s))

	CopyHeaders(s.Header, r)

	if err := ConfigureBackendURL(r, rl); err != nil {
		*r = *r.WithContext(context.WithValue(r.Context(), director, err))
		return
	}

	var en error // need to set it to error but with nil value
	*r = *r.WithContext(context.WithValue(r.Context(), director, en))
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
func EnrichRequestedURL(r *http.Request) {
	r.URL.Scheme = "http"
	r.URL.Host = r.Host
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		r.URL.Scheme = "https"
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
