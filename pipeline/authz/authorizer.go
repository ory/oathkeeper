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

package authz

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
)

var ErrAuthorizerNotEnabled = herodot.DefaultError{
	ErrorField:  "authorizer matching this route is misconfigured or disabled",
	CodeField:   http.StatusInternalServerError,
	StatusField: http.StatusText(http.StatusInternalServerError),
}

func NewErrAuthorizerNotEnabled(a Authorizer) *herodot.DefaultError {
	return ErrAuthorizerNotEnabled.WithTrace(errors.New("")).WithReasonf(`Authorizer "%s" is disabled per configuration.`, a.GetID())
}

func NewErrAuthorizerMisconfigured(a Authorizer, err error) *herodot.DefaultError {
	return ErrAuthorizerNotEnabled.WithTrace(err).WithReasonf(
		`Configuration for authorizer "%s" could not be validated: %s`,
		a.GetID(),
		err,
	)
}

type Authorizer interface {
	Authorize(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rule pipeline.Rule) error
	GetID() string
	Validate(config json.RawMessage) error
}
