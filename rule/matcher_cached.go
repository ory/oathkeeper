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
	"net/url"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pkg"
	"github.com/pkg/errors"
	"sync"
)

type CachedMatcher struct {
	Rules   map[string]Rule
	Manager Manager
	sync.RWMutex
}

func NewCachedMatcher(m Manager) *CachedMatcher {
	return &CachedMatcher{
		Manager: m,
		Rules:   map[string]Rule{},
	}
}

func (m *CachedMatcher) MatchRule(method string, u *url.URL) (*Rule, error) {
	m.RLock()
	defer m.RUnlock()
	var rules []Rule
	for _, rule := range m.Rules {
		if err := rule.IsMatching(method, u); err == nil {
			rules = append(rules, rule)
		}
	}

	if len(rules) == 0 {
		return nil, errors.WithStack(helper.ErrMatchesNoRule)
	} else if len(rules) != 1 {
		return nil, errors.WithStack(helper.ErrMatchesMoreThanOneRule)
	}

	return &rules[0], nil
}

func (m *CachedMatcher) Refresh() error {
	m.Lock()
	defer m.Unlock()

	rules, err := m.Manager.ListRules(pkg.RulesUpperLimit, 0)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, rule := range rules {
		m.Rules[rule.ID] = rule
	}
	return nil
}
