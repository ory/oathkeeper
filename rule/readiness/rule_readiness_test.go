package readiness

import (
	"testing"
)

func TestReadinessHealthChecker(t *testing.T) {
	t.Run("rule readiness probe", func(t *testing.T) {
		ruleReadinessProbe := NewReadinessHealthChecker()

		if name := ruleReadinessProbe.Name(); name != "rule-first-load" {
			t.Errorf("Name() did not returned expected name, name = %s", name)
			return
		}

		if err := ruleReadinessProbe.Validate(); err != ErrRuleNotYetLoaded {
			t.Errorf("Validate() did not returned expected error, error = %v", err)
			return
		}

		evtTypes := ruleReadinessProbe.EventTypes()
		if len(evtTypes) != 1 {
			t.Errorf("EventTypes() returned either 0 or multiple event type, evtTypes = %v", evtTypes)
			return
		}
		if _, ok := evtTypes[0].(*RuleLoadedEvent); !ok {
			t.Errorf("EventTypes() returned an unkown event type, evtTypes = %v", evtTypes)
			return
		}

		// Dispatch fake event
		ruleReadinessProbe.EventsReceiver(&RuleLoadedEvent{})

		if err := ruleReadinessProbe.Validate(); err != nil {
			t.Errorf("Validate() returned an error, error = %v", err)
			return
		}
	})
}
