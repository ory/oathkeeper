package health

import (
	"context"

	"github.com/ory/x/healthx"
	"github.com/pkg/errors"
)

type EventManager interface {
	Dispatch(event interface{})
	Watch(ctx context.Context)
	HealthxReadyCheckers() healthx.ReadyCheckers
}

var ErrEventTypeAlreadyRegistered = errors.New("event type already registered")
