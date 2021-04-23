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

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/driver/health"
	"github.com/ory/oathkeeper/helper"
	rulereadiness "github.com/ory/oathkeeper/rule/readiness"
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
	rules            []Rule
	matchingStrategy configuration.MatchingStrategy
	r                repositoryMemoryRegistry

	hem health.EventManager
}

// MatchingStrategy returns current MatchingStrategy.
func (m *RepositoryMemory) MatchingStrategy(_ context.Context) (configuration.MatchingStrategy, error) {
	m.RLock()
	defer m.RUnlock()
	return m.matchingStrategy, nil
}

// SetMatchingStrategy updates MatchingStrategy.
func (m *RepositoryMemory) SetMatchingStrategy(_ context.Context, ms configuration.MatchingStrategy) error {
	m.Lock()
	defer m.Unlock()
	m.matchingStrategy = ms
	return nil
}

func NewRepositoryMemory(r repositoryMemoryRegistry, hem health.EventManager) *RepositoryMemory {
	return &RepositoryMemory{
		r:     r,
		rules: make([]Rule, 0),
		hem:   hem,
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
			m.r.Logger().WithError(err).WithField("rule_id", check.ID).
				Errorf("A Rule uses a malformed configuration and all URLs matching this rule will not work. You should resolve this issue now.")
		}
	}

	m.Lock()
	m.rules = rules
	m.hem.Dispatch(&rulereadiness.RuleLoadedEvent{})
	m.Unlock()
	return nil
}

func (m *RepositoryMemory) Match(_ context.Context, method string, u *url.URL) (*Rule, error) {
	if u == nil {
		return nil, errors.WithStack(errors.New("nil URL provided"))
	}

	m.Lock()
	defer m.Unlock()

	var rules []Rule
	for k := range m.rules {
		r := &m.rules[k]
		if matched, err := r.IsMatching(m.matchingStrategy, method, u); err != nil {
			return nil, errors.WithStack(err)
		} else if matched {
			rules = append(rules, *r)
		}
		m.rules[k] = *r
	}

	if len(rules) == 0 {
		return nil, errors.WithStack(helper.ErrMatchesNoRule)
	} else if len(rules) != 1 {
		return nil, errors.WithStack(helper.ErrMatchesMoreThanOneRule)
	}

	return &rules[0], nil
}
