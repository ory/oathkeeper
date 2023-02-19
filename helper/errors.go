// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package helper

import (
	"net/http"

	"github.com/ory/herodot"
)

var (
	ErrForbidden = &herodot.DefaultError{
		ErrorField:  "Access credentials are not sufficient to access this resource",
		CodeField:   http.StatusForbidden,
		StatusField: http.StatusText(http.StatusForbidden),
	}
	ErrUnauthorized = &herodot.DefaultError{
		ErrorField:  "Access credentials are invalid",
		CodeField:   http.StatusUnauthorized,
		StatusField: http.StatusText(http.StatusUnauthorized),
	}
	ErrMatchesMoreThanOneRule = &herodot.DefaultError{
		ErrorField:  "Expected exactly one rule but found multiple rules",
		CodeField:   http.StatusInternalServerError,
		StatusField: http.StatusText(http.StatusInternalServerError),
	}
	ErrRuleFeatureDisabled = &herodot.DefaultError{
		ErrorField:  "The matched rule uses a feature which is not enabled in the server configuration",
		CodeField:   http.StatusInternalServerError,
		StatusField: http.StatusText(http.StatusInternalServerError),
	}
	// TODO: discuss the text and status code
	ErrNonRegexpMatchingStrategy = &herodot.DefaultError{
		ErrorField:  "The matched handler uses Regexp MatchingStrategy which is not selected in the configuration",
		CodeField:   http.StatusInternalServerError,
		StatusField: http.StatusText(http.StatusInternalServerError),
	}
	ErrMatchesNoRule = &herodot.DefaultError{
		ErrorField:  "Requested url does not match any rules",
		CodeField:   http.StatusNotFound,
		StatusField: http.StatusText(http.StatusNotFound),
	}
	ErrResourceNotFound = &herodot.DefaultError{
		ErrorField:  "The requested resource could not be found",
		CodeField:   http.StatusNotFound,
		StatusField: http.StatusText(http.StatusNotFound),
	}
	ErrResourceConflict = &herodot.DefaultError{
		ErrorField:  "The request could not be completed due to a conflict with the current state of the target resource",
		CodeField:   http.StatusConflict,
		StatusField: http.StatusText(http.StatusConflict),
	}
	ErrBadRequest = &herodot.DefaultError{
		ErrorField:  "The request is malformed or contains invalid data",
		CodeField:   http.StatusBadRequest,
		StatusField: http.StatusText(http.StatusBadRequest),
	}
	ErrUpstreamServiceNotAvailable = &herodot.DefaultError{
		ErrorField:  "The upstream service is not available",
		CodeField:   http.StatusServiceUnavailable,
		StatusField: http.StatusText(http.StatusServiceUnavailable),
	}
	ErrUpstreamServiceTimeout = &herodot.DefaultError{
		ErrorField:  "The upstream service is timing out",
		CodeField:   http.StatusGatewayTimeout,
		StatusField: http.StatusText(http.StatusGatewayTimeout),
	}
	ErrUpstreamServiceInternalServerError = &herodot.DefaultError{
		ErrorField:  "The upstream service encountered an unexpected error",
		CodeField:   http.StatusInternalServerError,
		StatusField: http.StatusText(http.StatusInternalServerError),
	}
	ErrUpstreamServiceNotFound = &herodot.DefaultError{
		ErrorField:  "Upstream service not found",
		CodeField:   http.StatusNotFound,
		StatusField: http.StatusText(http.StatusNotFound),
	}
)
