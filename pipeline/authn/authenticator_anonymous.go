// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"cmp"
	"encoding/json"
	"net/http"

	"github.com/ory/oathkeeper/pipeline"

	"github.com/pkg/errors"
)

type AuthenticatorAnonymous struct {
	d dependencies
}

type AuthenticatorAnonymousConfiguration struct {
	Subject string `json:"subject"`
}

func NewAuthenticatorAnonymous(d dependencies) *AuthenticatorAnonymous {
	return &AuthenticatorAnonymous{d: d}
}

func (a *AuthenticatorAnonymous) Validate(config json.RawMessage) error {
	if !a.d.Config().AuthenticatorIsEnabled(a.GetID()) {
		return NewErrAuthenticatorNotEnabled(a)
	}

	_, err := a.Config(config)
	return err
}

func (a *AuthenticatorAnonymous) GetID() string { return "anonymous" }

func (a *AuthenticatorAnonymous) Config(config json.RawMessage) (*AuthenticatorAnonymousConfiguration, error) {
	var c AuthenticatorAnonymousConfiguration
	if err := a.d.Config().AuthenticatorConfig(a.GetID(), config, &c); err != nil {
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

	session.Subject = cmp.Or(cf.Subject, "anonymous")

	return nil
}
