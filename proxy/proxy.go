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
	"strings"

	"net/url"

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/rsakey"
	"github.com/ory/oathkeeper/rule"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewProxy(handler *RequestHandler, logger logrus.FieldLogger, matcher rule.Matcher) *Proxy {
	if logger == nil {
		logger = logrus.New()
	}
	return &Proxy{
		Logger:         logger,
		Matcher:        matcher,
		RequestHandler: handler,
		H:              herodot.NewNegotiationHandler(logger),
	}
}

type Proxy struct {
	Logger         logrus.FieldLogger
	RequestHandler *RequestHandler
	KeyManager     rsakey.Manager
	Matcher        rule.Matcher
	H              herodot.Writer
}

type key int

const director key = 0

func (d *Proxy) RoundTrip(r *http.Request) (*http.Response, error) {
	rw := NewSimpleResponseWriter()

	if err, ok := r.Context().Value(director).(error); ok && err != nil {
		d.Logger.WithError(err).
			WithField("granted", false).
			WithField("access_url", r.URL.String()).
			Warn("Access request denied")

		d.H.WriteError(rw, r, err)

		return &http.Response{
			StatusCode: rw.code,
			Body:       ioutil.NopCloser(rw.buffer),
			Header:     rw.header,
		}, nil
	} else if err == nil {
		res, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			d.Logger.
				WithError(errors.WithStack(err)).
				WithField("granted", false).
				WithField("access_url", r.URL.String()).
				Warn("Access request denied because roundtrip failed")
			// don't need to return because covered in next line
		} else {
			d.Logger.
				WithField("granted", true).
				WithField("access_url", r.URL.String()).
				Warn("Access request granted")
		}

		return res, err
	}

	err := errors.New("Unable to type assert context")
	d.Logger.
		WithError(err).
		WithField("granted", false).
		WithField("access_url", r.URL.String()).
		Warn("Unable to type assert context")

	d.H.Write(rw, r, err)

	return &http.Response{
		StatusCode: rw.code,
		Body:       ioutil.NopCloser(rw.buffer),
		Header:     rw.header,
	}, nil
}

func (d *Proxy) Director(r *http.Request) {
	EnrichRequestedURL(r)
	rl, err := d.Matcher.MatchRule(r.Method, r.URL)
	if err != nil {
		*r = *r.WithContext(context.WithValue(r.Context(), director, err))
		return
	}

	if err := d.RequestHandler.HandleRequest(r, rl); err != nil {
		*r = *r.WithContext(context.WithValue(r.Context(), director, err))
		return
	}

	if err := configureBackendURL(r, rl); err != nil {
		*r = *r.WithContext(context.WithValue(r.Context(), director, err))
		return
	}

	var en error // need to set it to error but with nil value
	*r = *r.WithContext(context.WithValue(r.Context(), director, en))
}

// EnrichRequestedURL sets Scheme and Host values in a URL passed down by a http server. Per default, the URL
// does not contain host nor scheme values.
func EnrichRequestedURL(r *http.Request) {
	r.URL.Scheme = "http"
	r.URL.Host = r.Host
	if r.TLS != nil {
		r.URL.Scheme = "https"
	}
}

func configureBackendURL(r *http.Request, rl *rule.Rule) error {
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
