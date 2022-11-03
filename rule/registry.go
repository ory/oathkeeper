// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

type Registry interface {
	RuleValidator() Validator
	RuleFetcher() Fetcher
	RuleRepository() Repository
	RuleMatcher() Matcher
}
