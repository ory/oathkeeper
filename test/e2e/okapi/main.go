// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"net/http"
	"os"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"

	"github.com/ory/oathkeeper/x"
	"github.com/ory/x/urlx"
)

var (
	jwksProvider *jwks.Provider
	jwtValidator *validator.Validator
)

func init() {
	var err error
	u := x.ParseURLOrPanic(os.Getenv("OATHKEEPER_API"))
	jwksProvider = jwks.NewProvider(urlx.AppendPaths(u, "/.well-known/jwks.json"))
	jwtValidator, err = validator.New(
		jwksProvider.KeyFunc,
		validator.RS256,
		jwksProvider.IssuerURL.String(),
		[]string{jwksProvider.IssuerURL.String()},
	)
	if err != nil {
		panic(err)
	}
}

var jwtm = jwtmiddleware.New(jwtValidator.ValidateToken)

func main() {
	router := httprouter.New()

	router.GET("/jwt", jwtHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "6677"
	}

	n := negroni.Classic()
	n.UseHandler(router)
	n.UseFunc(func(_ http.ResponseWriter, _ *http.Request, next http.HandlerFunc) {
		jwtm.CheckJWT(next)
	})
	server := http.Server{ //nolint:gosec // test server without custom timeouts is acceptable
		Addr:    fmt.Sprintf(":%s", port),
		Handler: n,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func jwtHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	_, _ = w.Write([]byte("ok"))
}
