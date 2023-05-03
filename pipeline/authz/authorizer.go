// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
