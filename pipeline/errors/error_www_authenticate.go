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

package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/x"
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
		x.RegistryWriter
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
