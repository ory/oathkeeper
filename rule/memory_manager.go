package rule

type MemoryManager struct {
	Rules map[string]Rule
}

func NewMemoryManager() *MemoryManager {
	return &MemoryManager{Rules: map[string]Rule{}}
}

func (m *MemoryManager) ListRules() []Rule {
	rules := make([]Rule, len(m.Rules))
	i := 0
	for _, rule := range m.Rules {
		rules[i] = rule
		i++
	}
	return rules
}

func (m *MemoryManager) CreateRule(rule Rule) error {
	return m.UpdateRule(rule)
}

func (m *MemoryManager) UpdateRule(rule Rule) error {
	m.Rules[rule.ID] = rule
	return nil
}

func (m *MemoryManager) DeleteRule(id string) error {
	delete(m.Rules, id)
	return nil
}
