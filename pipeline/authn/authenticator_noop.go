// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/pipeline"
)

type AuthenticatorNoOp struct {
	d dependencies
}

func NewAuthenticatorNoOp(d dependencies) *AuthenticatorNoOp {
	return &AuthenticatorNoOp{d: d}
}

func (a *AuthenticatorNoOp) GetID() string { return "noop" }

func (a *AuthenticatorNoOp) Validate(config json.RawMessage) error {
	if !a.d.Config().AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	if err := a.d.Config().AuthenticatorConfig(a.GetID(), config, nil); err != nil {
		return NewErrAuthenticatorMisconfigured(a, err)
	}
	return nil
}

func (a *AuthenticatorNoOp) Authenticate(*http.Request, *AuthenticationSession, json.RawMessage, pipeline.Rule) error {
	return nil
}
