// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz

import (
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/helper"
)

type AuthorizerDeny struct {
	d configuration.Provider
}

func NewAuthorizerDeny(d configuration.Provider) *AuthorizerDeny {
	return &AuthorizerDeny{d: d}
}

func (a *AuthorizerDeny) GetID() string {
	return "deny"
}

func (a *AuthorizerDeny) Authorize(*http.Request, *authn.AuthenticationSession, json.RawMessage, pipeline.Rule) error {
	return errors.WithStack(helper.ErrForbidden())
}

func (a *AuthorizerDeny) Validate(config json.RawMessage) error {
	if !a.d.Config().AuthorizerIsEnabled(a.GetID()) {
		return NewErrAuthorizerNotEnabled(a)
	}

	if err := a.d.Config().AuthorizerConfig(a.GetID(), config, nil); err != nil {
		return NewErrAuthorizerMisconfigured(a, err)
	}
	return nil
}
