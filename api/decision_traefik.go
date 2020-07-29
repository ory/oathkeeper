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
	DecisionTraefikPath = "/decisions/traefik"
	TraefikProto        = "X-Forwarded-Proto"
	TraefikHost         = "X-Forwarded-Host"
	TraefikURI          = "X-Forwarded-Uri"
	TraefikMethod       = "X-Forwarded-Method"
)

type decisionTraefikHandlerRegistry interface {
	x.RegistryWriter
	x.RegistryLogger

	RuleMatcher() rule.Matcher
	ProxyRequestHandler() *proxy.RequestHandler
}

type DecisionTraefikHandler struct {
	r decisionTraefikHandlerRegistry
}

func NewDecisionTraefikerHandler(r decisionTraefikHandlerRegistry) *DecisionTraefikHandler {
	return &DecisionTraefikHandler{r: r}
}

func (h *DecisionTraefikHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if r.Method == "GET" && len(r.URL.Path) >= len(DecisionTraefikPath) && r.URL.Path[:len(DecisionTraefikPath)] == DecisionTraefikPath {
		r.URL.Scheme = "http"
		r.URL.Host = r.Host
		if r.TLS != nil {
			r.URL.Scheme = "https"
		}

		h.decisionTraefik(w, r)
	} else {
		next(w, r)
	}
}

// swagger:route GET /decisions/traefik api makeTraefikDecision
//
// Access Control Decision Traefik API
//
// This endpoint mirrors the proxy capability of ORY Oathkeeper's proxy functionality but instead of forwarding the
// request to the upstream server, returns 200 (request should be allowed), 401 (unauthorized), or 403 (forbidden)
// status codes. This endpoint can be used to integrate with the Traefik proxy.
//
//     Schemes: http, https
//
//     Responses:
//       200: emptyResponse
//       401: genericError
//       403: genericError
//       404: genericError
//       500: genericError
func (h *DecisionTraefikHandler) decisionTraefik(w http.ResponseWriter, r *http.Request) {
	urlToMatch := url.URL{
		Scheme: r.Header.Get(TraefikProto),
		Host:   r.Header.Get(TraefikHost),
		Path:   r.Header.Get(TraefikURI),
	}
	methodToMatch := r.Header.Get(TraefikMethod)

	rl, err := h.r.RuleMatcher().Match(r.Context(), methodToMatch, &urlToMatch)
	if err != nil {
		h.r.Logger().WithError(err).
			WithField("granted", false).
			WithField("access_url", urlToMatch.String()).
			Warn("Access request denied")
		h.r.ProxyRequestHandler().HandleError(w, r, rl, err)
		return
	}

	s, err := h.r.ProxyRequestHandler().HandleRequest(r, rl)
	if err != nil {
		h.r.Logger().WithError(err).
			WithField("granted", false).
			WithField("access_url", urlToMatch.String()).
			Warn("Access request denied")
		h.r.ProxyRequestHandler().HandleError(w, r, rl, err)
		return
	}

	h.r.Logger().
		WithField("granted", true).
		WithField("access_url", urlToMatch.String()).
		Warn("Access request granted")

	for k := range s.Header {
		// Avoid copying the original Content-Length header from the client
		if k == "content-length" {
			continue
		}

		w.Header().Set(k, s.Header.Get(k))
	}

	w.WriteHeader(http.StatusOK)
}
