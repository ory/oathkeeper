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
	"github.com/ory/oathkeeper/helper"
	"github.com/pkg/errors"
)

type MemoryManager struct {
	Rules map[string]Rule
}

func NewMemoryManager() *MemoryManager {
	return &MemoryManager{Rules: map[string]Rule{}}
}

func (m *MemoryManager) ListRules() ([]Rule, error) {
	rules := make([]Rule, len(m.Rules))
	i := 0
	for _, rule := range m.Rules {
		rules[i] = rule
		i++
	}
	return rules, nil
}

func (m *MemoryManager) GetRule(id string) (*Rule, error) {
	if rule, ok := m.Rules[id]; !ok {
		return nil, errors.WithStack(helper.ErrResourceNotFound)
	} else {
		return &rule, nil
	}
}

func (m *MemoryManager) CreateRule(rule *Rule) error {
	return m.UpdateRule(rule)
}

func (m *MemoryManager) UpdateRule(rule *Rule) error {
	m.Rules[rule.ID] = *rule
	return nil
}

func (m *MemoryManager) DeleteRule(id string) error {
	delete(m.Rules, id)
	return nil
}
