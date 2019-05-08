package api

import (
	"net/http"
	"net/url"

	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/x"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/square/go-jose.v2"
)

type credentialHandlerRegistry interface {
	x.RegistryWriter
	credentials.FetcherRegistry
}

type CredentialsHandler struct {
	c configuration.Provider
	r credentialHandlerRegistry
}

func NewCredentialHandler(c configuration.Provider, r credentialHandlerRegistry) *CredentialsHandler {
	return &CredentialsHandler{c: c, r: r}
}

func (h *CredentialsHandler) SetRoutes(r *x.RouterAPI) {
	r.GET("/.well-known/jwks.json", h.wellKnown)
}

// swagger:route GET /.well-known/jwks.json api getWellKnownJSONWebKeys
//
// Lists cryptographic keys
//
// This endpoint returns cryptographic keys that are required to, for example, verify signatures of ID Tokens.
//
//     Produces:
//     - application/json
//
//     Schemes: http, https
//
//     Responses:
//       200: jsonWebKeySet
//       500: genericError
func (h *CredentialsHandler) wellKnown(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sets, err := h.r.CredentialsFetcher().ResolveSets(r.Context(), []url.URL{
		*h.c.MutatorIDTokenJWKSURL(),
	})
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}

	keys := make([]jose.JSONWebKey, 0)
	for _, set := range sets {
		for _, key := range set.Keys {
			if p := key.Public(); p.Key != nil {
				keys = append(keys, p)
			}
		}
	}

	h.r.Writer().Write(w, r, &jose.JSONWebKeySet{Keys: keys})
}
