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
	"net/url"
	"sync"

	"github.com/pkg/errors"

	"github.com/ory/x/viperx"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/x"

	"github.com/ory/x/pagination"
)

var _ Repository = new(RepositoryMemory)

type repositoryMemoryRegistry interface {
	RuleValidator() Validator
	x.RegistryLogger
}

type RepositoryMemory struct {
	sync.RWMutex
	rules []Rule
	r     repositoryMemoryRegistry
}

func NewRepositoryMemory(r repositoryMemoryRegistry) *RepositoryMemory {
	return &RepositoryMemory{
		r:     r,
		rules: make([]Rule, 0),
	}
}

// WithRules sets rules without validation. For testing only.
func (m *RepositoryMemory) WithRules(rules []Rule) {
	m.Lock()
	m.rules = rules
	m.Unlock()
}

func (m *RepositoryMemory) Count(ctx context.Context) (int, error) {
	m.RLock()
	defer m.RUnlock()

	return len(m.rules), nil
}

func (m *RepositoryMemory) List(ctx context.Context, limit, offset int) ([]Rule, error) {
	m.RLock()
	defer m.RUnlock()

	start, end := pagination.Index(limit, offset, len(m.rules))
	return m.rules[start:end], nil
}

func (m *RepositoryMemory) Get(ctx context.Context, id string) (*Rule, error) {
	m.RLock()
	defer m.RUnlock()

	for _, r := range m.rules {
		if r.ID == id {
			return &r, nil
		}
	}

	return nil, errors.WithStack(helper.ErrResourceNotFound)
}

func (m *RepositoryMemory) Set(ctx context.Context, rules []Rule) error {
	for _, check := range rules {
		if err := m.r.RuleValidator().Validate(&check); err != nil {
			viperx.LoggerWithValidationErrorFields(m.r.Logger(), err).WithError(err).
				Errorf("A rule uses a malformed configuration and all URLs matching this rule will not work. You should resolve this issue now.")
		}
	}

	m.Lock()
	m.rules = rules
	m.Unlock()
	return nil
}

func (m *RepositoryMemory) Match(ctx context.Context, method string, u *url.URL) (*Rule, error) {
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
