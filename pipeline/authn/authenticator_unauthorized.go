/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

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

func (a *AuthenticatorUnauthorized) Authenticate(r *http.Request, config json.RawMessage, _ pipeline.Rule) (*AuthenticationSession, error) {
	return nil, errors.WithStack(helper.ErrUnauthorized)
}
