package rule

type Registry interface {
	RuleValidator() Validator
	RuleRepository() Repository
	RuleMatcher() Matcher
}
