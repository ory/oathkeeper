// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"net/http"
	"net/url"

	"github.com/tidwall/gjson"

	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline/mutate"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/httpx"
	"github.com/ory/x/urlx"

	"github.com/go-jose/go-jose/v3"
	"github.com/julienschmidt/httprouter"
)

const (
	CredentialsPath = "/.well-known/jwks.json" //nolint:gosec // public metadata location, not a credential literal
)

type credentialHandlerRegistry interface {
	httpx.WriterProvider
	credentials.FetcherRegistry
	rule.Registry
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
// # Lists Cryptographic Keys
//
// This endpoint returns cryptographic keys that are required to, for example, verify signatures of ID Tokens.
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Responses:
//	  200: jsonWebKeySet
//	  500: genericError
func (h *CredentialsHandler) wellKnown(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	urls, err := h.jwksURLs()
	if err != nil {
		h.r.Writer().WriteError(w, r, err)
		return
	}
	sets, err := h.r.CredentialsFetcher().ResolveSets(r.Context(), urls)
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

func (h *CredentialsHandler) jwksURLs() ([]url.URL, error) {
	t := map[string]bool{}
	for _, u := range h.c.JSONWebKeyURLs() {
		t[u] = true
	}

	rules, err := h.r.RuleRepository().List(context.Background(), 2147483647, 0)
	if err != nil {
		return nil, err
	}
	for _, r := range rules {
		for _, m := range r.Mutators {
			if m.Handler == new(mutate.MutatorIDToken).GetID() {
				u := gjson.GetBytes(m.Config, "jwks_url").String()
				if len(u) > 0 {
					t[u] = true
				}
			}
		}
	}

	result := make([]url.URL, len(t))
	i := 0
	for u := range t {
		uu, err := urlx.Parse(u)
		if err != nil {
			return nil, err
		}
		result[i] = *uu
		i++
	}

	return result, nil
}
