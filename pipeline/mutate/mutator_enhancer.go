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
	"github.com/ory/x/httpx"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/pipeline/authn"
)

var ErrMalformedResponseFromUpstreamAPI = errors.New("The call to an external API returned an invalid JSON object")
var ErrMissingAPIURL = errors.New("Missing URL in mutator configuration")
var ErrInvalidAPIURL = errors.New("Invalid URL in mutator configuration")
var ErrNon200ResponseFromAPI = errors.New("The call to an external API returned a non-200 HTTP response")

type MutatorEnhancer struct {
	c      configuration.Provider
	client *http.Client
}

func NewMutatorEnhancer(c configuration.Provider) *MutatorEnhancer {
	return &MutatorEnhancer{c: c, client: httpx.NewResilientClientLatencyToleranceSmall(nil)}
}

func (a *MutatorEnhancer) GetID() string {
	return "enhancer"
}

func (a *MutatorEnhancer) Mutate(r *http.Request, session *authn.AuthenticationSession, config json.RawMessage, _ pipeline.Rule) (http.Header, error) {
	return nil, nil
}

func (a *MutatorEnhancer) Validate() error {
	if !a.c.MutatorEnhancerIsEnabled() {
		return errors.WithStack(ErrMutatorNotEnabled.WithReasonf(`Mutator "%s" is disabled per configuration.`, a.GetID()))
	}
	return nil
}
