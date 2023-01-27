// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
