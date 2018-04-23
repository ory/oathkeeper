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

	"github.com/ory/ladon/compiler"
	"github.com/ory/oathkeeper/pkg"
	"github.com/ory/oathkeeper/sdk/go/oathkeeper"
	"github.com/pkg/errors"
	//"net/url"
	"net/url"
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
		if r.RequiredScopes == nil {
			r.RequiredScopes = []string{}
		}

		if r.MatchesMethods == nil {
			r.MatchesMethods = []string{}
		}

		matches, err := compiler.CompileRegex(r.MatchesUrl, '<', '>')
		if err != nil {
			return errors.WithStack(err)
		}

		parsed, err := url.Parse(r.Upstream.Url)
		if err != nil {
			return errors.WithStack(err)
		}

		m.Rules[r.Id] = Rule{
			ID:                 r.Id,
			Description:        r.Description,
			MatchesMethods:     r.MatchesMethods,
			Mode:               r.Mode,
			MatchesURL:         r.MatchesUrl,
			MatchesURLCompiled: matches,
			RequiredAction:     r.RequiredAction,
			RequiredResource:   r.RequiredResource,
			RequiredScopes:     r.RequiredScopes,
			Upstream: &Upstream{
				URL:          r.Upstream.Url,
				URLParsed:    parsed,
				PreserveHost: r.Upstream.PreserveHost,
				StripPath:    r.Upstream.StripPath,
			},
		}
	}

	return nil
}
