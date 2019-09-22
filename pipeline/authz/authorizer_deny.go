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
	c configuration.Provider
}

func NewAuthorizerDeny(c configuration.Provider) *AuthorizerDeny {
	return &AuthorizerDeny{c: c}
}

func (a *AuthorizerDeny) GetID() string {
	return "deny"
}

func (a *AuthorizerDeny) Authorize(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) error {
	return errors.WithStack(helper.ErrForbidden)
}

func (a *AuthorizerDeny) Validate(config json.RawMessage) error {
	if !a.c.AuthorizerIsEnabled(a.GetID()) {
		return NewErrAuthorizerNotEnabled(a)
	}

	if err := a.c.AuthorizerConfig(a.GetID(), config, nil); err != nil {
		return NewErrAuthorizerMisconfigured(a, err)
	}
	return nil
}
