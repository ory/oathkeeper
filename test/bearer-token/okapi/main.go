// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()

	router.POST("/test", test)
	router.GET("/session", session)

	port := os.Getenv("PORT")
	if port == "" {
		port = "6662"
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func test(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	if len(b) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func session(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.WriteHeader(http.StatusOK)
}
