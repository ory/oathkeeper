// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
)

type AuthenticatorDelegate struct {
	c configuration.Provider
}

func NewAuthenticatorDelegate(c configuration.Provider) *AuthenticatorDelegate {
	return &AuthenticatorDelegate{c: c}
}

func (a *AuthenticatorDelegate) GetID() string {
	return "delegate"
}

func (a *AuthenticatorDelegate) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	if err := a.c.AuthenticatorConfig(a.GetID(), config, nil); err != nil {
		return NewErrAuthenticatorMisconfigured(a, err)
	}
	return nil
}

func (a *AuthenticatorDelegate) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	return ErrAuthenticatorDelegate
}
