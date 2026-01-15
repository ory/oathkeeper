// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"

	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/x"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/x/httpx"
	"github.com/ory/x/pagination"
)

const (
	RulesPath = "/rules"
)

type RuleHandler struct {
	r ruleHandlerRegistry
}

type ruleHandlerRegistry interface {
	httpx.WriterProvider
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
// # List All Rules
//
// This method returns an array of all rules that are stored in the backend. This is useful if you want to get a full
// view of what rules you have currently in place.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: rules
//	  500: genericError
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
// # Retrieve a Rule
//
// Use this method to retrieve a rule from the storage. If it does not exist you will receive a 404 error.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: rule
//	  404: genericError
//	  500: genericError
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
