package health

import (
	"context"
	"testing"
	"time"

	rulereadiness "github.com/ory/oathkeeper/rule/readiness"
)

func TestNewDefaultHealthEventManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Run("health event manager", func(t *testing.T) {
		ruleReadinessProbe := rulereadiness.NewReadinessHealthChecker()
		hem := NewDefaultHealthEventManager()

		if err := hem.AddListener(ruleReadinessProbe); err != nil {
			t.Errorf("AddListener() error = %v", err)
			return
		}

		if err := hem.AddListener(ruleReadinessProbe); err == nil {
			t.Errorf("AddListener() was able to register twice the same listener")
			return
		}

		// Dispatch event without watching
		hem.Dispatch(&rulereadiness.RuleLoadedEvent{})

		// Watching for incoming events
		go func() {
			hem.Watch(ctx)
		}()

		// Waiting for watcher to be ready
		time.Sleep(100*time.Millisecond)
		// Dispatch event
		hem.Dispatch(&rulereadiness.RuleLoadedEvent{})
		// Wait for event propagation
		time.Sleep(100*time.Millisecond)

		if err := ruleReadinessProbe.Validate(); err != nil {
			t.Errorf("Validate() returned an unexpected error, err = %v", err)
			return
		}
		cancel()
	})
}
