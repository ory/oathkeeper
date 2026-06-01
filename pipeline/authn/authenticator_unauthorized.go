// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/pipeline"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/helper"
)

type AuthenticatorUnauthorized struct {
	d dependencies
}

func NewAuthenticatorUnauthorized(d dependencies) *AuthenticatorUnauthorized {
	return &AuthenticatorUnauthorized{d: d}
}

func (a *AuthenticatorUnauthorized) Validate(config json.RawMessage) error {
	if !a.d.Config().AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	if err := a.d.Config().AuthenticatorConfig(a.GetID(), config, nil); err != nil {
		return NewErrAuthenticatorMisconfigured(a, err)
	}
	return nil
}

func (a *AuthenticatorUnauthorized) GetID() string { return "unauthorized" }

func (a *AuthenticatorUnauthorized) Authenticate(*http.Request, *AuthenticationSession, json.RawMessage, pipeline.Rule) error {
	return errors.WithStack(helper.ErrUnauthorized())
}
