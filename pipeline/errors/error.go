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

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/pipeline"
)

type Handler interface {
	GetID() string
	Handle(w http.ResponseWriter, r *http.Request, config json.RawMessage, _ pipeline.Rule, err error) error
	Validate(config json.RawMessage) error
}

var ErrErrorHandlerNotEnabled = herodot.DefaultError{
	ErrorField:  "error handler matching this route is misconfigured or disabled",
	CodeField:   http.StatusInternalServerError,
	StatusField: http.StatusText(http.StatusInternalServerError),
}

var ErrHandlerNotResponsible = errors.New("error handler not responsible for this request")

func NewErrErrorHandlerNotEnabled(a Handler) *herodot.DefaultError {
	return ErrErrorHandlerNotEnabled.WithTrace(errors.New("")).WithReasonf(`Error handler "%s" is disabled per configuration.`, a.GetID())
}

func NewErrErrorHandlerMisconfigured(a Handler, err error) *herodot.DefaultError {
	return ErrErrorHandlerNotEnabled.WithTrace(err).WithReasonf(
		`Configuration for error handler "%s" could not be validated: %s`,
		a.GetID(),
		err,
	)
}
