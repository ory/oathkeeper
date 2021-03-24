package health

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const mockReadinessProbeName = "mock-readiness-probe"

type (
	mockReadinessProbe struct {
		hasReceivedEvent bool
		testData         string
	}
	mockReadinessEvent struct {
		testData string
	}
)

func (m *mockReadinessProbe) ID() string {
	return mockReadinessProbeName
}

func (m *mockReadinessProbe) Validate() error {
	return nil
}

func (m *mockReadinessProbe) EventTypes() []ReadinessProbeEvent {
	return []ReadinessProbeEvent{&mockReadinessEvent{}}
}

func (m *mockReadinessProbe) EventsReceiver(evt ReadinessProbeEvent) {
	switch castedEvent := evt.(type) {
	case *mockReadinessEvent:
		m.hasReceivedEvent = true
		m.testData = castedEvent.testData
	}
}

func (m *mockReadinessEvent) ReadinessProbeListenerID() string {
	return mockReadinessProbeName
}

func TestNewDefaultHealthEventManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	t.Run("health event manager", func(t *testing.T) {
		readinessProbe := &mockReadinessProbe{}

		// Create a new default health event manager with twice same probe
		_, err := NewDefaultHealthEventManager(readinessProbe, readinessProbe)
		require.Error(t, err)

		// Create a new default health event manager
		hem, err := NewDefaultHealthEventManager(readinessProbe)
		require.NoError(t, err)

		// Test healthx ready checkers generation
		checkers := hem.HealthxReadyCheckers()
		require.Len(t, checkers, 1)
		_, ok := checkers[readinessProbe.ID()]
		require.True(t, ok, "health checker was not found")

		// Readiness probe must be empty before event dispatch
		require.False(t, readinessProbe.hasReceivedEvent)
		require.Equal(t, readinessProbe.testData, "")

		// Nil events should be ignored
		hem.Dispatch(nil)
		require.False(t, readinessProbe.hasReceivedEvent)

		// Dispatch event without watching (should not block)
		const testData = "a sample string that will be passed along the event"
		hem.Dispatch(&mockReadinessEvent{
			testData: testData,
		})

		// Watching for incoming events
		hem.Watch(ctx)

		// Waiting for watcher to be ready, then verify the event has been received
		time.Sleep(100 * time.Millisecond)
		require.True(t, readinessProbe.hasReceivedEvent)
		require.Equal(t, readinessProbe.testData, testData)

		// Reset probe
		readinessProbe.hasReceivedEvent = false
		readinessProbe.testData = ""

		// Dispatch a new event
		hem.Dispatch(&mockReadinessEvent{
			testData: testData,
		})

		// Wait for event propagation, then verify the event has been received
		time.Sleep(100 * time.Millisecond)
		require.True(t, readinessProbe.hasReceivedEvent)
		require.Equal(t, readinessProbe.testData, testData)
		cancel()
	})
}
