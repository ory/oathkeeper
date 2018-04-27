package extauth

import (
	"github.com/julienschmidt/httprouter"
	"github.com/ory/oathkeeper/evaluator"
	"github.com/ory/oathkeeper/helper"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
)

type Handler struct {
	Evaluator evaluator.Evaluator
}

func (h *Handler) SetRoutes(r *httprouter.Router) {
	r.GET("/extauth", h.Extauth)
}

// swagger:route GET /extauth
//
// Checks if a token is valid and if the token subject is allowed to perform an action on a resource.
// This endpoint requires a token, a scope, a resource name, an action name and a context.
// If a token is expired/invalid, has not been granted the requested scope or the subject is not allowed to
// perform the action on the resource, this endpoint returns a 403 response.
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
//       200: Ok
//       401: genericError
//       403: genericError
//       500: genericError
func (h *Handler) Extauth(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	u, err := url.Parse(r.Header.Get("x-original-url"))
	if err != nil || u == nil || u.String() == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	method := r.Header.Get("x-original-method")
	if method == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r.URL = u
	r.Host = u.Host
	r.Method = method

	access, err := h.Evaluator.EvaluateAccessRequest(r)
	if err != nil {
		switch errors.Cause(err) {
		case helper.ErrForbidden:
			w.WriteHeader(http.StatusForbidden)
		case helper.ErrMissingBearerToken:
			w.WriteHeader(http.StatusUnauthorized)
		case helper.ErrUnauthorized:
			w.WriteHeader(http.StatusUnauthorized)
		case helper.ErrMatchesNoRule:
			w.WriteHeader(http.StatusNotFound)
		case helper.ErrMatchesMoreThanOneRule:
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	if access.Disabled {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}
