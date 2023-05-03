// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz

import (
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
)

type AuthorizerAllow struct {
	c configuration.Provider
}

func NewAuthorizerAllow(c configuration.Provider) *AuthorizerAllow {
	return &AuthorizerAllow{c: c}
}

func (a *AuthorizerAllow) GetID() string {
	return "allow"
}

func (a *AuthorizerAllow) Authorize(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	return nil
}

func (a *AuthorizerAllow) Validate(config json.RawMessage) error {
	if !a.c.AuthorizerIsEnabled(a.GetID()) {
		return NewErrAuthorizerNotEnabled(a)
	}

	if err := a.c.AuthorizerConfig(a.GetID(), config, nil); err != nil {
		return NewErrAuthorizerMisconfigured(a, err)
	}
	return nil
}
