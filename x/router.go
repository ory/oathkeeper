// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"github.com/julienschmidt/httprouter"

	"github.com/ory/x/serverx"
)

type RouterAPI struct {
	*httprouter.Router
}

func NewAPIRouter() *RouterAPI {
	router := httprouter.New()
	router.NotFound = serverx.DefaultNotFoundHandler
	return &RouterAPI{
		Router: router,
	}
}
