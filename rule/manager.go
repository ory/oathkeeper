package rule

type Manager interface {
	ListRules() ([]Rule, error)
	AddRule(rule Rule) error
	RemoveRule(id string) error
	UpdateRule(rule Rule) error
}
