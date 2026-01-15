// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/x/httpx"
)

var _ Handler = new(ErrorWWWAuthenticate)

type (
	ErrorWWWAuthenticateConfig struct {
		Realm string `json:"realm"`
	}
	ErrorWWWAuthenticate struct {
		c configuration.Provider
		d ErrorWWWAuthenticateDependencies
	}
	ErrorWWWAuthenticateDependencies interface {
		httpx.WriterProvider
	}
)

func NewErrorWWWAuthenticate(
	c configuration.Provider,
	d ErrorWWWAuthenticateDependencies,
) *ErrorWWWAuthenticate {
	return &ErrorWWWAuthenticate{c: c, d: d}
}

func (a *ErrorWWWAuthenticate) Handle(w http.ResponseWriter, r *http.Request, config json.RawMessage, _ pipeline.Rule, _ error) error {
	c, err := a.Config(config)
	if err != nil {
		return err
	}

	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm=%s`, c.Realm))
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	return nil
}

func (a *ErrorWWWAuthenticate) Validate(config json.RawMessage) error {
	if !a.c.ErrorHandlerIsEnabled(a.GetID()) {
		return NewErrErrorHandlerNotEnabled(a)
	}
	_, err := a.Config(config)
	return err
}

func (a *ErrorWWWAuthenticate) Config(config json.RawMessage) (*ErrorWWWAuthenticateConfig, error) {
	var c ErrorWWWAuthenticateConfig
	if err := a.c.ErrorHandlerConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrErrorHandlerMisconfigured(a, err)
	}

	if c.Realm == "" {
		c.Realm = "Please authenticate."
	}

	return &c, nil
}

func (a *ErrorWWWAuthenticate) GetID() string {
	return "www_authenticate"
}
