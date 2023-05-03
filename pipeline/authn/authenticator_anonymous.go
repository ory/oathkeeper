// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"encoding/json"
	"net/http"

	"github.com/ory/x/stringsx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"

	"github.com/pkg/errors"
)

type AuthenticatorAnonymous struct {
	c configuration.Provider
}

type AuthenticatorAnonymousConfiguration struct {
	Subject string `json:"subject"`
}

func NewAuthenticatorAnonymous(c configuration.Provider) *AuthenticatorAnonymous {
	return &AuthenticatorAnonymous{
		c: c,
	}
}

func (a *AuthenticatorAnonymous) Validate(config json.RawMessage) error {
	if !a.c.AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *AuthenticatorAnonymous) GetID() string {
	return "anonymous"
}

func (a *AuthenticatorAnonymous) Config(config json.RawMessage) (*AuthenticatorAnonymousConfiguration, error) {
	var c AuthenticatorAnonymousConfiguration
	if err := a.c.AuthenticatorConfig(a.GetID(), config, &c); err != nil {
		return nil, NewErrAuthenticatorMisconfigured(a, err)
	}

	return &c, nil
}

func (a *AuthenticatorAnonymous) Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	if len(r.Header.Get("Authorization")) != 0 {
		return errors.WithStack(ErrAuthenticatorNotResponsible)
	}

	cf, err := a.Config(config)
	if err != nil {
		return err
	}

	session.Subject = stringsx.Coalesce(cf.Subject, "anonymous")

	return nil
}
