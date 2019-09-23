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

package mutate

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
)

var ErrMutatorNotEnabled = herodot.DefaultError{
	ErrorField:  "mutator matching this route is misconfigured or disabled",
	CodeField:   http.StatusInternalServerError,
	StatusField: http.StatusText(http.StatusInternalServerError),
}

func NewErrMutatorNotEnabled(a Mutator) *herodot.DefaultError {
	return ErrMutatorNotEnabled.WithTrace(errors.New("")).WithReasonf(`Mutator "%s" is disabled per configuration.`, a.GetID())
}

func NewErrMutatorMisconfigured(a Mutator, err error) *herodot.DefaultError {
	return ErrMutatorNotEnabled.WithTrace(err).WithReasonf(
		`Configuration for mutator "%s" could not be validated: %s`,
		a.GetID(),
		err,
	)
}

type Mutator interface {
	Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error
	GetID() string
	Validate(config json.RawMessage) error
}
