// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/helper"
)

type AuthenticatorUnauthorized struct {
	c configuration.Provider
}

func NewAuthenticatorUnauthorized(c configuration.Provider) *AuthenticatorUnauthorized {
	return &AuthenticatorUnauthorized{c: c}
}

func (a *AuthenticatorUnauthorized) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	if err := a.c.AuthenticatorConfig(a.GetID(), config, nil); err != nil {
		return NewErrAuthenticatorMisconfigured(a, err)
	}
	return nil
}

func (a *AuthenticatorUnauthorized) GetID() string {
	return "unauthorized"
}

func (a *AuthenticatorUnauthorized) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	return errors.WithStack(helper.ErrUnauthorized)
}
