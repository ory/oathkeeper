package readiness

import (
	"github.com/pkg/errors"
)

type (
	RuleReadinessChecker struct {
		hasFirstRuleLoad bool
	}

	RuleLoadedEvent struct {}
)

func NewReadinessHealthChecker() *RuleReadinessChecker {
	return &RuleReadinessChecker{
		hasFirstRuleLoad: false,
	}
}

func (r *RuleReadinessChecker) Name() string {
	return "rule-first-load"
}

func (r *RuleReadinessChecker) Validate() error {
	if !r.hasFirstRuleLoad {
		return errors.New("rules have not been loaded yet")
	}
	return nil
}

func (r *RuleReadinessChecker) EventTypes() []interface{} {
	return []interface{}{&RuleLoadedEvent{}}
}

func (r *RuleReadinessChecker) EventsReceiver(event interface{}) {
	switch event.(type) {
	case *RuleLoadedEvent:
		r.hasFirstRuleLoad = true
	}
}
