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

package proxy

import (
	"encoding/json"
	"github.com/ory/oathkeeper/driver/configuration"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rule"
)

type AuthorizerDeny struct {
	c configuration.Provider
}

func NewAuthorizerDeny(c configuration.Provider) *AuthorizerDeny {
	return &AuthorizerDeny{c: c}
}

func (a *AuthorizerDeny) GetID() string {
	return "deny"
}

func (a *AuthorizerDeny) Authorize(r *http.Request, session *AuthenticationSession, config json.RawMessage, rl *rule.Rule) error {
	return errors.WithStack(helper.ErrForbidden)
}

func (a *AuthorizerDeny) Validate() error {
	if !a.c.AuthorizerDenyIsEnabled() {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf("Authorizer % is disabled per configuration.", a.GetID()))
	}

	return nil
}
