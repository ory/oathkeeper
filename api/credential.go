// Copyright 2021 Ory GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"github.com/ory/x/urlx"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/square/go-jose.v2"
)

const (
	CredentialsPath = "/.well-known/jwks.json"
)

type credentialHandlerRegistry interface {
	x.RegistryWriter
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
