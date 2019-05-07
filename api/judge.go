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

	"github.com/ory/oathkeeper/x"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/proxy"
	"github.com/ory/oathkeeper/rule"
)

const (
	JudgePath = "/judge"
)

type judgeHandlerRegistry interface {
	x.RegistryWriter
}

type JudgeHandler struct {
	Logger         logrus.FieldLogger
	RequestHandler *proxy.RequestHandler
	Matcher        rule.Matcher
	H              herodot.Writer
	Router         *httprouter.Router
}

func NewJudgeHandler(handler *proxy.RequestHandler, logger logrus.FieldLogger, matcher rule.Matcher, router *httprouter.Router) *JudgeHandler {
	if logger == nil {
		logger = logrus.New()
	}
	return &JudgeHandler{
		Logger:         logger,
		Matcher:        matcher,
		RequestHandler: handler,
		H:              herodot.NewNegotiationHandler(logger),
		Router:         router,
	}
}

func (h *JudgeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) >= len(JudgePath) && r.URL.Path[:len(JudgePath)] == JudgePath {
		r.URL.Scheme = "http"
		r.URL.Host = r.Host
		if r.TLS != nil {
			r.URL.Scheme = "https"
		}
		r.URL.Path = r.URL.Path[len(JudgePath):]

		h.judge(w, r)
	} else {
		h.Router.ServeHTTP(w, r)
	}
}

// swagger:route GET /judge judge judge
//
// Judge if a request should be allowed or not
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
func (h *JudgeHandler) judge(w http.ResponseWriter, r *http.Request) {
	rl, err := h.Matcher.Match(r.Context(), r.Method, r.URL)
	if err != nil {
		h.Logger.WithError(err).
			WithField("granted", false).
			WithField("access_url", r.URL.String()).
			Warn("Access request denied")
		h.H.WriteError(w, r, err)
		return
	}

	headers, err := h.RequestHandler.HandleRequest(r, rl)
	if err != nil {
		h.Logger.WithError(err).
			WithField("granted", false).
			WithField("access_url", r.URL.String()).
			Warn("Access request denied")
		h.H.WriteError(w, r, err)
		return
	}

	h.Logger.
		WithField("granted", true).
		WithField("access_url", r.URL.String()).
		Warn("Access request granted")

	for k := range headers {
		w.Header().Set(k, headers.Get(k))
	}

	w.WriteHeader(http.StatusOK)
}
