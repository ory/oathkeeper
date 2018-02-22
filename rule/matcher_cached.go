// Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package rule

import (
	"net/url"

	"github.com/ory/oathkeeper/helper"
	"github.com/pkg/errors"
)

type CachedMatcher struct {
	Rules   []Rule
	Manager Manager
}

func (m *CachedMatcher) MatchRule(method string, u *url.URL) (*Rule, error) {
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
	rules, err := m.Manager.ListRules()
	if err != nil {
		return errors.WithStack(err)
	}

	m.Rules = rules
	return nil
}
