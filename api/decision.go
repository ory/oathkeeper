// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"cmp"
	"net/http"
	"strings"

	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/httpx"
	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rule"
)

const (
	DecisionPath = "/decisions"

	xForwardedMethod = "X-Forwarded-Method"
	xForwardedProto  = "X-Forwarded-Proto"
	xForwardedHost   = "X-Forwarded-Host"
	xForwardedUri    = "X-Forwarded-Uri"
)

type decisionHandlerRegistry interface {
	httpx.WriterProvider
	logrusx.Provider

	RuleMatcher() rule.Matcher
	ProxyRequestHandler() proxy.RequestHandler
}

type DecisionHandler struct {
	r decisionHandlerRegistry
}

func NewJudgeHandler(r decisionHandlerRegistry) *DecisionHandler {
	return &DecisionHandler{r: r}
}

func (h *DecisionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if len(r.URL.Path) >= len(DecisionPath) && r.URL.Path[:len(DecisionPath)] == DecisionPath {
		r.Method = cmp.Or(r.Header.Get(xForwardedMethod), r.Method)
		r.URL.Scheme = cmp.Or(r.Header.Get(xForwardedProto),
			x.IfThenElseString(r.TLS != nil, "https", "http"))
		r.URL.Host = cmp.Or(r.Header.Get(xForwardedHost), r.Host)
		r.URL.Path = cmp.Or(strings.SplitN(r.Header.Get(xForwardedUri), "?", 2)[0], r.URL.Path[len(DecisionPath):])

		h.decisions(w, r)
	} else {
		next(w, r)
	}
}

// swagger:route GET /decisions api decisions
//
// # Access Control Decision API
//
// > This endpoint works with all HTTP Methods (GET, POST, PUT, ...) and matches every path prefixed with /decisions.
//
// This endpoint mirrors the proxy capability of ORY Oathkeeper's proxy functionality but instead of forwarding the
// request to the upstream server, returns 200 (request should be allowed), 401 (unauthorized), or 403 (forbidden)
// status codes. This endpoint can be used to integrate with other API Proxies like Ambassador, Kong, Envoy, and many more.
//
//	Schemes: http, https
//
//	Responses:
//	  200: emptyResponse
//	  401: genericError
//	  403: genericError
//	  404: genericError
//	  500: genericError
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

	rl, err := h.r.RuleMatcher().Match(r.Context(), r.Method, r.URL, rule.ProtocolHTTP)
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
