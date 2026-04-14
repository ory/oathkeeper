// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"errors"
	"net/http"

	pkgerrors "github.com/pkg/errors"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
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
		httpx.WriterProvider
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

	var errH *helper.ErrWithHeaders
	if errors.As(handleError, &errH) {
		for key, vals := range errH.Headers {
			for _, v := range vals {
				w.Header().Set(key, v)
			}
		}
	}

	if !c.Verbose {
		wrapped := pkgerrors.WithStack(handleError)
		var sc statusCoder
		if errors.As(handleError, &sc) {
			switch sc.StatusCode() {
			case http.StatusInternalServerError:
				handleError = herodot.ErrInternalServerError().WithWrap(wrapped)
			case http.StatusForbidden:
				handleError = herodot.ErrForbidden().WithWrap(wrapped)
			case http.StatusNotFound:
				handleError = herodot.ErrNotFound().WithWrap(wrapped)
			case http.StatusUnauthorized:
				handleError = herodot.ErrUnauthorized().WithWrap(wrapped)
			case http.StatusBadRequest:
				handleError = herodot.ErrBadRequest().WithWrap(wrapped)
			case http.StatusTooManyRequests:
				handleError = helper.ErrTooManyRequests().WithWrap(wrapped)
			case http.StatusUnsupportedMediaType:
				handleError = herodot.ErrUnsupportedMediaType().WithWrap(wrapped)
			case http.StatusConflict:
				handleError = herodot.ErrConflict().WithWrap(wrapped)
			}
		} else {
			handleError = herodot.ErrInternalServerError().WithWrap(wrapped)
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
