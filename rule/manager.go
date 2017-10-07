package rule

type Manager interface {
	ListRules() ([]Rule, error)
	CreateRule(*Rule) error
	GetRule(id string) (*Rule, error)
	DeleteRule(id string) error
	UpdateRule(*Rule) error
}
