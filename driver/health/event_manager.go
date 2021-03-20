package health

import (
	"context"
)

type EventManager interface {
	Dispatch(event interface{})
	AddListener(listener Readiness) error
	Watch(ctx context.Context)
}
