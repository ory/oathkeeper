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
