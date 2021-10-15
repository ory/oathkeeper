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

	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/x"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/x/pagination"
)

const (
	RulesPath = "/rules"
)

type RuleHandler struct {
	r ruleHandlerRegistry
}

type ruleHandlerRegistry interface {
	x.RegistryWriter
	rule.Registry
}

func NewRuleHandler(r ruleHandlerRegistry) *RuleHandler {
	return &RuleHandler{r: r}
}

func (h *RuleHandler) SetRoutes(r *x.RouterAPI) {
	r.GET(RulesPath, h.listRules)
	r.GET(RulesPath+"/:id", h.getRules)
}

// swagger:route GET /rules api listRules
//
// List All Rules
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
//       500: genericError
func (h *RuleHandler) listRules(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	limit, offset := pagination.Parse(r, 50, 0, 500)
	rules, err := h.r.RuleRepository().List(r.Context(), limit, offset)
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	if rules == nil {
		rules = make([]rule.Rule, 0)
	}

	h.r.Writer().Write(w, r, rules)
}

// swagger:route GET /rules/{id} api getRule
//
// Retrieve a Rule
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
//       404: genericError
//       500: genericError
func (h *RuleHandler) getRules(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	rl, err := h.r.RuleRepository().Get(r.Context(), ps.ByName("id"))
	if errors.Cause(err) == helper.ErrResourceNotFound {
		h.r.Writer().WriteErrorCode(w, r, http.StatusNotFound, err)
		return
	} else if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	h.r.Writer().Write(w, r, rl)
}
