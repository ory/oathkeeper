// Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package helper

import "github.com/pkg/errors"

var (
	ErrMissingBearerToken     = errors.New("This action requires authorization but no bearer token was given")
	ErrForbidden              = errors.New("Access credentials are not sufficient to access this resource")
	ErrUnauthorized           = errors.New("Access credentials are either expired or missing a scope")
	ErrMatchesMoreThanOneRule = errors.New("Expected exactly one rule but found multiple rules")
	ErrMatchesNoRule          = errors.New("Requested url does not match any rules")
	ErrResourceNotFound       = errors.New("The requested resource could not be found")
)
