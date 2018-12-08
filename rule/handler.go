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

package rule

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pkg"
	"github.com/ory/x/pagination"
)

type Handler struct {
	H herodot.Writer
	M Manager
	V func(*Rule) error
}

func NewHandler(
	h herodot.Writer,
	m Manager,
	v func(*Rule) error,
) *Handler {
	return &Handler{
		H: h,
		M: m,
		V: v,
	}
}

func (h *Handler) SetRoutes(r *httprouter.Router) {
	r.GET("/rules", h.List)
	r.POST("/rules", h.Create)
	r.GET("/rules/:id", h.Get)
	r.PUT("/rules/:id", h.Update)
	r.DELETE("/rules/:id", h.Delete)
}

// swagger:route POST /rules rule createRule
//
// Create a rule
//
// This method allows creation of rules. If a rule id exists, you will receive an error.
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       201: rule
//       401: genericError
//       403: genericError
//       500: genericError
func (h *Handler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rule, err := h.decodeRule(w, r)
	if err != nil {
		h.H.WriteError(w, r, err)
		return
	}

	if rule.ID == "" {
		rule.ID = uuid.New()
	}

	if err := h.M.CreateRule(rule); err != nil {
		h.H.WriteError(w, r, err)
		return
	}

	h.H.WriteCreated(w, r, "/rules/"+rule.ID, rule)
}

// swagger:route GET /rules rule listRules
//
// List all rules
//
// This method returns an array of all rules that are stored in the backend. This is useful if you want to get a full
// view of what rules you have currently in place.
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: rules
//       401: genericError
//       403: genericError
//       500: genericError
func (h *Handler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	limit, offset := pagination.Parse(r, 100, 0, pkg.RulesUpperLimit)
	rules, err := h.M.ListRules(limit, offset)
	if err != nil {
		h.H.WriteError(w, r, err)
		return
	}

	h.H.Write(w, r, rules)
}

// swagger:route GET /rules/{id} rule getRule
//
// Retrieve a rule
//
// Use this method to retrieve a rule from the storage. If it does not exist you will receive a 404 error.
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: rule
//       401: genericError
//       403: genericError
//       404: genericError
//       500: genericError
func (h *Handler) Get(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rule, err := h.M.GetRule(ps.ByName("id"))
	if errors.Cause(err) == helper.ErrResourceNotFound {
		h.H.WriteErrorCode(w, r, http.StatusNotFound, err)
		return
	} else if err != nil {
		h.H.WriteError(w, r, err)
		return
	}

	h.H.Write(w, r, rule)
}

// swagger:route PUT /rules/{id} rule updateRule
//
// Update a rule
//
// Use this method to update a rule. Keep in mind that you need to send the full rule payload as this endpoint does
// not support patching.
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: rule
//       401: genericError
//       403: genericError
//       404: genericError
//       500: genericError
func (h *Handler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rule, err := h.decodeRule(w, r)
	if err != nil {
		h.H.WriteError(w, r, err)
		return
	}

	rule.ID = ps.ByName("id")
	if err := h.M.UpdateRule(rule); err != nil {
		h.H.WriteError(w, r, err)
		return
	}

	h.H.Write(w, r, rule)
}

// swagger:route DELETE /rules/{id} rule deleteRule
//
// Delete a rule
//
// Use this endpoint to delete a rule.
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       204: emptyResponse
//       401: genericError
//       403: genericError
//       404: genericError
//       500: genericError
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if err := h.M.DeleteRule(ps.ByName("id")); err != nil {
		h.H.WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) decodeRule(w http.ResponseWriter, r *http.Request) (*Rule, error) {
	rule := NewRule()

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(rule); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := h.V(rule); err != nil {
		return nil, err
	}

	return rule, nil
}
