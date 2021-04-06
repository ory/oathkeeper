package readiness

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestReadinessHealthChecker(t *testing.T) {
	t.Run("rule readiness probe", func(t *testing.T) {
		ruleReadinessProbe := NewReadinessHealthChecker()
		ruleLoadedEvent := RuleLoadedEvent{}

		assert.Equal(t, ruleReadinessProbe.ID(), ProbeName)
		assert.Equal(t, ruleLoadedEvent.ReadinessProbeListenerID(), ProbeName)

		assert.True(t, errors.Is(ruleReadinessProbe.Validate(), ErrRuleNotYetLoaded))

		evtTypes := ruleReadinessProbe.EventTypes()
		assert.Len(t, evtTypes, 1)
		_, ok := evtTypes[0].(*RuleLoadedEvent)
		assert.True(t, ok, "actual type %T", evtTypes[0])

		// Dispatch fake event
		ruleReadinessProbe.EventsReceiver(&RuleLoadedEvent{})

		assert.NoError(t, ruleReadinessProbe.Validate())
	})
}
