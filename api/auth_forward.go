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

package api

import (
	"net/http"
	"net/url"

	"github.com/ory/oathkeeper/x"

	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rule"
)

const (
	AuthForwardPath   = "/auth_forward"
	xForwardedURI     = "X-Forwarded-Uri"
	xForwardedMethod  = "X-Forwarded-Method"
)

type authForwardHandlerRegistry interface {
	x.RegistryWriter
	x.RegistryLogger

	RuleMatcher() rule.Matcher
	ProxyRequestHandler() *proxy.RequestHandler
}

type AuthForwardHandler struct {
	r authForwardHandlerRegistry
}

func NewAuthForwarderHandler(r authForwardHandlerRegistry) *AuthForwardHandler {
	return &AuthForwardHandler{r: r}
}

func (h *AuthForwardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if len(r.URL.Path) >= len(AuthForwardPath) && r.URL.Path[:len(AuthForwardPath)] == AuthForwardPath {
		r.URL.Scheme = "http"
		r.URL.Host = r.Host
		if r.TLS != nil {
			r.URL.Scheme = "https"
		}
		r.URL.Path = r.URL.Path[len(AuthForwardPath):]

		h.authForwards(w, r)
	} else {
		next(w, r)
	}
}

// swagger:route GET /authForwards api authForwards
//
// Access Control AuthForward API
//
// > This endpoint works with all HTTP Methods (GET, POST, PUT, ...) and matches every path prefixed with /authForward.
//
// This endpoint mirrors the proxy capability of ORY Oathkeeper's proxy functionality but instead of forwarding the
// request to the upstream server, returns 200 (request should be allowed), 401 (unauthorized), or 403 (forbidden)
// status codes. This endpoint can be used to integrate with other API Proxies like Ambassador, Kong, Envoy, and many more.
//
//     Schemes: http, https
//
//     Responses:
//       200: emptyResponse
//       401: genericError
//       403: genericError
//       404: genericError
//       500: genericError
func (h *AuthForwardHandler) authForwards(w http.ResponseWriter, r *http.Request) {
	uriToMatch, err := url.Parse(r.Header.Get(xForwardedURI))
	methodToMatch := r.Header.Get(xForwardedMethod)

	rl, err := h.r.RuleMatcher().Match(r.Context(), methodToMatch, uriToMatch)
	if err != nil {
		h.r.Logger().WithError(err).
			WithField("granted", false).
			WithField("access_url", uriToMatch.String()).
			Warn("Access request denied")
		h.r.Writer().WriteError(w, r, err)
		return
	}

	headers, err := h.r.ProxyRequestHandler().HandleRequest(r, rl)
	if err != nil {
		h.r.Logger().WithError(err).
			WithField("granted", false).
			WithField("access_url", uriToMatch.String()).
			Warn("Access request denied")
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Logger().
		WithField("granted", true).
		WithField("access_url", uriToMatch.String()).
		Warn("Access request granted")

	for k := range headers {
		w.Header().Set(k, headers.Get(k))
	}

	w.WriteHeader(http.StatusOK)
}
