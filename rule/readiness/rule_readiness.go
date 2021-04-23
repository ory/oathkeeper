package readiness

import (
	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/driver/health"
)

type (
	RuleReadinessChecker struct {
		hasFirstRuleLoad bool
	}

	RuleLoadedEvent struct{}
)

const ProbeName = "rule-first-load"

var ErrRuleNotYetLoaded = errors.New("rules have not been loaded yet")

func NewReadinessHealthChecker() *RuleReadinessChecker {
	return &RuleReadinessChecker{
		hasFirstRuleLoad: false,
	}
}

func (r *RuleReadinessChecker) ID() string {
	return ProbeName
}

func (r *RuleReadinessChecker) Validate() error {
	if !r.hasFirstRuleLoad {
		return errors.WithStack(ErrRuleNotYetLoaded)
	}
	return nil
}

func (r *RuleReadinessChecker) EventTypes() []health.ReadinessProbeEvent {
	return []health.ReadinessProbeEvent{&RuleLoadedEvent{}}
}

func (r *RuleReadinessChecker) EventsReceiver(event health.ReadinessProbeEvent) {
	switch event.(type) {
	case *RuleLoadedEvent:
		r.hasFirstRuleLoad = true
	}
}

func (r *RuleLoadedEvent) ReadinessProbeListenerID() string {
	return ProbeName
}
