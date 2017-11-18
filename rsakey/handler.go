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
	r.GET("/keys/id-token.public", h.GetPublicKey)
}

func (h *Handler) GetPublicKey(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	key, err := h.M.PublicKey()
	if err != nil {
		h.H.WriteError(w, r, err)
		return
	}

	jwk := &jose.JSONWebKey{
		Key:       key,
		KeyID:     "id-token.public",
		Algorithm: h.M.Algorithm(),
	}

	h.H.Write(w, r, jwk)
}
