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
	"net/http"

	"github.com/ory/herodot"
	"github.com/ory/x/errorsx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/x"
)

var _ Handler = new(ErrorJSON)

type (
	ErrorJSONConfig struct {
		Verbose bool `json:"verbose"`
	}
	ErrorJSON struct {
		c configuration.Provider
		d errorJSONDependencies
	}
	errorJSONDependencies interface {
		x.RegistryWriter
	}
)

func NewErrorJSON(
	c configuration.Provider,
	d errorJSONDependencies,
) *ErrorJSON {
	return &ErrorJSON{c: c, d: d}
}

func (a *ErrorJSON) Handle(w http.ResponseWriter, r *http.Request, config json.RawMessage, _ pipeline.Rule, handleError error) error {
	c, err := a.Config(config)
	if err != nil {
		return err
	}

	if !c.Verbose {
		if sc, ok := errorsx.Cause(handleError).(statusCoder); ok {
			switch sc.StatusCode() {
			case http.StatusInternalServerError:
				handleError = herodot.ErrInternalServerError.WithTrace(handleError)
			case http.StatusForbidden:
				handleError = herodot.ErrForbidden.WithTrace(handleError)
			case http.StatusNotFound:
				handleError = herodot.ErrNotFound.WithTrace(handleError)
			case http.StatusUnauthorized:
				handleError = herodot.ErrUnauthorized.WithTrace(handleError)
			case http.StatusBadRequest:
				handleError = herodot.ErrBadRequest.WithTrace(handleError)
			case http.StatusUnsupportedMediaType:
				handleError = herodot.ErrUnsupportedMediaType.WithTrace(handleError)
			case http.StatusConflict:
				handleError = herodot.ErrConflict.WithTrace(handleError)
			}
		} else {
			handleError = herodot.ErrInternalServerError.WithTrace(handleError)
		}
	}

	a.d.Writer().WriteError(w, r, handleError)
	return nil
}

func (a *ErrorJSON) Validate(config json.RawMessage) error {
	if !a.c.ErrorHandlerIsEnabled(a.GetID()) {
		return NewErrErrorHandlerNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *ErrorJSON) Config(config json.RawMessage) (*ErrorJSONConfig, error) {
	var c ErrorJSONConfig
	if err := a.c.ErrorHandlerConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrErrorHandlerMisconfigured(a, err)
	}

	return &c, nil
}

func (a *ErrorJSON) GetID() string {
	return "json"
}
