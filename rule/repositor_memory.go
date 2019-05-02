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
	"context"
	"github.com/pkg/errors"
	"net/url"
	"sync"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/x/pagination"
)

type MemoryManager struct {
	sync.RWMutex
	rules []Rule
}

func NewMemoryManager() *MemoryManager {
	return &MemoryManager{rules: []Rule{}}
}

func (m *MemoryManager) Count(ctx context.Context) (int, error) {
	return len(m.rules), nil
}

func (m *MemoryManager) List(limit, offset int) ([]Rule, error) {
	start, end := pagination.Index(limit, offset, len(m.rules))
	return m.rules[start:end], nil
}

func (m *MemoryManager) Get(id string) (*Rule, error) {
	for _, r := range m.rules {
		if r.ID == id {
			return &r, nil
		}
	}

	return nil, errors.WithStack(helper.ErrResourceNotFound)
}

func (m *MemoryManager) Upsert(rule *Rule) error {
	for k, r := range m.rules {
		if r.ID == rule.ID {
			m.rules[k] = *rule
			return nil
		}
	}

	m.rules = append(m.rules, *rule)
	return nil
}

func (m *MemoryManager) Delete(id string) error {
	for k, r := range m.rules {
		if r.ID == id {
			m.rules = append(m.rules[:k], m.rules[k+1:]...)
			return nil
		}
	}

	return errors.WithStack(helper.ErrResourceNotFound)
}

func (m *MemoryManager) Match(method string, u *url.URL) (*Rule, error) {
	m.RLock()
	defer m.RUnlock()
	var rules []Rule
	for _, rule := range m.rules {
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
