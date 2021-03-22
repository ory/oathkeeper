package health

import (
	"context"
	"reflect"

	"github.com/ory/x/healthx"
	"github.com/pkg/errors"
)

type DefaultHealthEventManager struct {
	evtChan                chan interface{}
	listeners              []Readiness
	listenerEventTypeCache map[reflect.Type]Readiness
}

func NewDefaultHealthEventManager(listeners ...Readiness) (*DefaultHealthEventManager, error) {
	var listenerEventTypeCache = make(map[reflect.Type]Readiness)
	for _, listener := range listeners {
		for _, evtType := range listener.EventTypes() {
			evtTypeVal := reflect.TypeOf(evtType)
			if _, ok := listenerEventTypeCache[evtTypeVal]; ok {
				return nil, errors.WithStack(ErrEventTypeAlreadyRegistered)
			}
			listenerEventTypeCache[reflect.TypeOf(evtType)] = listener
		}
	}
	return &DefaultHealthEventManager{
		evtChan:                make(chan interface{}),
		listeners:              listeners,
		listenerEventTypeCache: listenerEventTypeCache,
	}, nil
}

func (h *DefaultHealthEventManager) Dispatch(event interface{}) {
	go func() {
		h.evtChan <- event
	}()
}

func (h *DefaultHealthEventManager) Watch(ctx context.Context) {
	go func() {
		for {
			var evt interface{}
			select {
			case evt = <-h.evtChan:
			case <-ctx.Done():
				return
			}
			if listener, ok := h.listenerEventTypeCache[reflect.TypeOf(evt)]; ok {
				listener.EventsReceiver(evt)
			}
		}
	}()
}

func (h *DefaultHealthEventManager) HealthxReadyCheckers() healthx.ReadyCheckers {
	var checkers = make(healthx.ReadyCheckers)
	for _, listener := range h.listeners {
		checkers[listener.Name()] = listener.Validate
	}
	return checkers
}
