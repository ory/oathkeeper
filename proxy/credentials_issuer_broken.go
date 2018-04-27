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
	"net/http"
	"encoding/json"
	"errors"
)

type CredentialsIssuerBroken struct{}

func NewCredentialsIssuerBroken() *CredentialsIssuerBroken {
	return new(CredentialsIssuerBroken)
}

func (a *CredentialsIssuerBroken) GetID() string {
	return "broken"
}

func (a *CredentialsIssuerBroken) Issue(r *http.Request, session *AuthenticationSession, config json.RawMessage) error {
	return errors.New("forced denial of credentials")
}
