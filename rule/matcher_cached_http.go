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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/pkg"
	"github.com/ory/x/urlx"
)

type HTTPMatcher struct {
	u *url.URL
	c *http.Client
	*CachedMatcher
}

func NewHTTPMatcher(u *url.URL) *HTTPMatcher {
	return &HTTPMatcher{
		u: u,
		c: &http.Client{Timeout: time.Second * 5},
		CachedMatcher: &CachedMatcher{
			Rules: map[string]Rule{},
		},
	}
}

func (m *HTTPMatcher) Refresh() error {
	from := urlx.CopyWithQuery(urlx.AppendPaths(m.u, "/rules"), url.Values{"limit": {fmt.Sprintf("%d", pkg.RulesUpperLimit)}})
	res, err := http.DefaultClient.Get(from.String())
	if err != nil {
		return errors.WithStack(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("unable to fetch rules from backend got status code %d but expected %d when calling %s", res.StatusCode, http.StatusOK, from)
	}

	var rules []Rule
	if err := json.NewDecoder(res.Body).Decode(&rules); err != nil {
		return errors.WithStack(err)
	}

	m.Lock()
	defer m.Unlock()
	inserted := map[string]bool{}
	for _, r := range rules {
		if len(r.Match.Methods) == 0 {
			r.Match.Methods = []string{}
		}

		if len(r.Authenticators) == 0 {
			r.Authenticators = []RuleHandler{}
		}

		inserted[r.ID] = true
		m.Rules[r.ID] = r
	}

	for _, r := range m.Rules {
		if _, ok := inserted[r.ID]; !ok {
			delete(m.Rules, r.ID)
		}
	}

	return nil
}
