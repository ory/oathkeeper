package health

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	rulereadiness "github.com/ory/oathkeeper/rule/readiness"
)

func TestNewDefaultHealthEventManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Run("health event manager", func(t *testing.T) {
		ruleReadinessProbe := rulereadiness.NewReadinessHealthChecker()

		// Create a new default health event manager with twice same probe
		_, err := NewDefaultHealthEventManager(ruleReadinessProbe, ruleReadinessProbe)
		require.Error(t, err)

		// Create a new default health event manager
		hem, err := NewDefaultHealthEventManager(ruleReadinessProbe)
		require.NoError(t, err)

		// Test healthx ready checkers generation
		checkers := hem.HealthxReadyCheckers()
		require.Len(t, checkers, 1)
		_, ok := checkers[ruleReadinessProbe.Name()]
		require.True(t, ok, "health checker was not found")

		// Rule readiness probe must return an error before event dispatch
		require.True(t, errors.Is(ruleReadinessProbe.Validate(), rulereadiness.ErrRuleNotYetLoaded))

		// Dispatch event without watching (should not block)
		hem.Dispatch(&rulereadiness.RuleLoadedEvent{})

		// Watching for incoming events
		hem.Watch(ctx)

		// Waiting for watcher to be ready
		time.Sleep(100 * time.Millisecond)
		// Dispatch event
		hem.Dispatch(&rulereadiness.RuleLoadedEvent{})
		// Wait for event propagation
		time.Sleep(100 * time.Millisecond)

		require.NoError(t, ruleReadinessProbe.Validate())
		cancel()
	})
}
