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

func (a *AuthenticatorDelegate) Validate(_ json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}
	return nil
}

func (a *AuthenticatorDelegate) Authenticate(r *http.Request, _ *AuthenticationSession, _ json.RawMessage, _ pipeline.Rule) error {
	return ErrAuthenticatorDelegate
}
