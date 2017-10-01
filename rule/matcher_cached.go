package rule

import (
	"github.com/pkg/errors"
	"net/url"
)

type CachedMatcher struct {
	Rules   []Rule
	Manager Manager
}

func (m *CachedMatcher) MatchRules(method string, u *url.URL) (rules []Rule, err error) {
	for _, rule := range m.Rules {
		if err := rule.MatchesURL(method, u); err == nil {
			rules = append(rules, rule)
		}
	}

	if len(rules) == 0 {
		return nil, errors.Errorf("Unable to finde rule matching %s:%s", method, u.String())
	}

	return rules, nil
}

func (m *CachedMatcher) Refresh() error {
	rules, err := m.Manager.ListRules()
	if err != nil {
		return errors.WithStack(err)
	}

	m.Rules = rules
	return nil
}
