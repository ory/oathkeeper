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

package rsakey

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ory/herodot"
	"gopkg.in/square/go-jose.v2"
)

type Handler struct {
	H herodot.Writer
	M Manager
}

func (h *Handler) SetRoutes(r *httprouter.Router) {
	r.GET("/.well-known/jwks.json", h.WellKnown)
}

// swagger:route GET /.well-known/jwks.json getWellKnown
//
// Returns well known keys
//
// This endpoint returns public keys for validating the ID tokens issued by ORY Oathkeeper.
//
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: jsonWebKeySet
//       401: genericError
//       403: genericError
func (h *Handler) WellKnown(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	key, err := h.M.PublicKey()
	if err != nil {
		h.H.WriteError(w, r, err)
		return
	}

	h.H.Write(w, r, &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{{
			Key:       key,
			Use:       "sig",
			KeyID:     h.M.PublicKeyID(),
			Algorithm: h.M.Algorithm(),
		}},
	})
}
