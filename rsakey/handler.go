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

func (h *Handler) WellKnown(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	key, err := h.M.PublicKey()
	if err != nil {
		h.H.WriteError(w, r, err)
		return
	}

	h.H.Write(w, r, &jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{{
			Key:       key,
			KeyID:     h.M.PublicKeyID(),
			Algorithm: h.M.Algorithm(),
		}},
	})
}
