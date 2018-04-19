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

package rule

import (
	"net/http"

	"github.com/ory/oathkeeper/pkg"
	"github.com/ory/oathkeeper/sdk/go/oathkeeper"
	"github.com/pkg/errors"
)

type HTTPMatcher struct {
	O oathkeeper.SDK
	*CachedMatcher
}

func NewHTTPMatcher(o oathkeeper.SDK) *HTTPMatcher {
	return &HTTPMatcher{
		O: o,
		CachedMatcher: &CachedMatcher{
			Rules: map[string]Rule{},
		},
	}
}

func (m *HTTPMatcher) Refresh() error {
	rules, response, err := m.O.ListRules(pkg.RulesUpperLimit, 0)
	if err != nil {
		return errors.WithStack(err)
	}
	if response.StatusCode != http.StatusOK {
		return errors.Errorf("Unable to fetch rules from backend, got status code %d but expected %s", response.StatusCode, http.StatusOK)
	}

	for _, r := range rules {
		m.Rules[r.Id] = Rule{
			ID:               r.Id,
			Description:      r.Description,
			MatchesMethods:   r.MatchesMethods,
			Mode:             r.Mode,
			MatchesURL:       r.MatchesUrl,
			RequiredAction:   r.RequiredAction,
			RequiredResource: r.RequiredResource,
			RequiredScopes:   r.RequiredScopes,
		}
	}

	return nil
}
