package rule

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/ory/herodot"
	"github.com/ory/ladon/compiler"
	"github.com/ory/oathkeeper/helper"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

type Handler struct {
	H herodot.Writer
	M Manager
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
	rule, err := decodeRule(w, r)
	if err != nil {
		h.H.WriteError(w, r, errors.WithStack(err))
		return
	}

	if rule.ID == "" {
		rule.ID = uuid.New()
	}

	if err := h.M.CreateRule(rule); err != nil {
		h.H.WriteError(w, r, err)
		return
	}

	h.H.WriteCreated(w, r, "/rules/"+rule.ID, encodeRule(rule))
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
	rules, err := h.M.ListRules()
	if err != nil {
		h.H.WriteError(w, r, err)
		return
	}

	var encodedRules []jsonRule = make([]jsonRule, len(rules))
	for k, rule := range rules {
		encodedRules[k] = *encodeRule(&rule)
	}

	h.H.Write(w, r, encodedRules)
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

	h.H.Write(w, r, encodeRule(rule))
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
	rule, err := decodeRule(w, r)
	if err != nil {
		h.H.WriteError(w, r, errors.WithStack(err))
		return
	}

	rule.ID = ps.ByName("id")
	if err := h.M.UpdateRule(rule); err != nil {
		h.H.WriteError(w, r, err)
		return
	}

	h.H.Write(w, r, encodeRule(rule))
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

func decodeRule(w http.ResponseWriter, r *http.Request) (*Rule, error) {
	rule := jsonRule{
		MatchesMethods: []string{},
		RequiredScopes: []string{},
	}
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		return nil, err
	}

	return toRule(&rule)
}

func toRule(rule *jsonRule) (*Rule, error) {
	exp, err := compiler.CompileRegex(rule.MatchesURL, '<', '>')
	if err != nil {
		return nil, err
	}

	return &Rule{
		ID:                            rule.ID,
		MatchesURLCompiled:            exp,
		MatchesURL:                    rule.MatchesURL,
		MatchesMethods:                rule.MatchesMethods,
		RequiredScopes:                rule.RequiredScopes,
		RequiredAction:                rule.RequiredAction,
		RequiredResource:              rule.RequiredResource,
		AllowAnonymousModeEnabled:     rule.AllowAnonymousModeEnabled,
		PassThroughModeEnabled:        rule.PassThroughModeEnabled,
		BasicAuthorizationModeEnabled: rule.BasicAuthorizationModeEnabled,
		Description:                   rule.Description,
	}, nil
}

func encodeRule(r *Rule) *jsonRule {
	return &jsonRule{
		ID:                            r.ID,
		MatchesURL:                    r.MatchesURL,
		MatchesMethods:                r.MatchesMethods,
		RequiredScopes:                r.RequiredScopes,
		RequiredAction:                r.RequiredAction,
		RequiredResource:              r.RequiredResource,
		PassThroughModeEnabled:        r.PassThroughModeEnabled,
		BasicAuthorizationModeEnabled: r.BasicAuthorizationModeEnabled,
		AllowAnonymousModeEnabled:     r.AllowAnonymousModeEnabled,
		Description:                   r.Description,
	}
}
