// Copyright © 2022 Ory Corp

package rule

type Registry interface {
	RuleValidator() Validator
	RuleFetcher() Fetcher
	RuleRepository() Repository
	RuleMatcher() Matcher
}
