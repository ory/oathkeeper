package health

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ory/x/healthx"
)

type EventManager interface {
	Dispatch(event ReadinessProbeEvent)
	Watch(ctx context.Context)
	HealthxReadyCheckers() healthx.ReadyCheckers
}

var ErrEventTypeAlreadyRegistered = errors.New("event type already registered")
