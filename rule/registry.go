package rule

type Registry interface {
	RuleValidator() ValidatorDefault
	RuleManager() Repository
}
