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
	"strings"

	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/x"

	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rule"
)

const (
	DecisionPath = "/decisions"
)

type decisionHandlerRegistry interface {
	x.RegistryWriter
	x.RegistryLogger

	RuleMatcher() rule.Matcher
	ProxyRequestHandler() *proxy.RequestHandler
}

type DecisionHandler struct {
	r decisionHandlerRegistry
}

func NewJudgeHandler(r decisionHandlerRegistry) *DecisionHandler {
	return &DecisionHandler{r: r}
}

func (h *DecisionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if len(r.URL.Path) >= len(DecisionPath) && r.URL.Path[:len(DecisionPath)] == DecisionPath {
		r.URL.Scheme = "http"
		r.URL.Host = r.Host
		if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
			r.URL.Scheme = "https"
		}
		r.URL.Path = r.URL.Path[len(DecisionPath):]

		h.decisions(w, r)
	} else {
		next(w, r)
	}
}

// swagger:route GET /decisions api decisions
//
// Access Control Decision API
//
// > This endpoint works with all HTTP Methods (GET, POST, PUT, ...) and matches every path prefixed with /decision.
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
func (h *DecisionHandler) decisions(w http.ResponseWriter, r *http.Request) {
	fields := map[string]interface{}{
		"http_method":     r.Method,
		"http_url":        r.URL.String(),
		"http_host":       r.Host,
		"http_user_agent": r.UserAgent(),
	}

	if sess, ok := r.Context().Value(proxy.ContextKeySession).(*authn.AuthenticationSession); ok {
		fields["subject"] = sess.Subject
	}

	rl, err := h.r.RuleMatcher().Match(r.Context(), r.Method, r.URL)
	if err != nil {
		h.r.Logger().WithError(err).
			WithFields(fields).
			WithField("granted", false).
			Warn("Access request denied")
		h.r.ProxyRequestHandler().HandleError(w, r, rl, err)
		return
	}

	s, err := h.r.ProxyRequestHandler().HandleRequest(r, rl)
	if err != nil {
		h.r.Logger().WithError(err).
			WithFields(fields).
			WithField("granted", false).
			Info("Access request denied")

		h.r.ProxyRequestHandler().HandleError(w, r, rl, err)
		return
	}

	h.r.Logger().
		WithFields(fields).
		WithField("granted", true).
		Info("Access request granted")

	for k := range s.Header {
		// Avoid copying the original Content-Length header from the client
		if strings.ToLower(k) == "content-length" {
			continue
		}

		w.Header().Set(k, s.Header.Get(k))
	}

	w.WriteHeader(http.StatusOK)
}
