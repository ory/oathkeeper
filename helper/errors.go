/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

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
)
