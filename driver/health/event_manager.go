package health

import (
	"context"
)

type EventManager interface {
	Dispatch(event interface{})
	AddListener(listener Readiness)
	Watch(ctx context.Context)
}
