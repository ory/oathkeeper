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

package mutate

import (
	"encoding/json"
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline/authn"
	"github.com/pkg/errors"
	"net/http"

	"github.com/ory/oathkeeper/rule"
)

type MutatorNoop struct{c configuration.Provider}

func NewCredentialsIssuerNoOp(c configuration.Provider) *MutatorNoop {
	return &MutatorNoop{c:c}
}

func (a *MutatorNoop) GetID() string {
	return "noop"
}

func (a *MutatorNoop) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, rl *rule.Rule) (http.Header, error) {
	return r.Header, nil
}

func (a *MutatorNoop) Validate() error {
	if !a.c.MutatorNoopIsEnabled() {
		return errors.WithStack(authn.ErrAuthenticatorNotEnabled.WithReasonf("Mutator % is disabled per configuration.", a.GetID()))
	}

	return nil
}
