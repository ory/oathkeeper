package health

import (
	"context"
	"errors"
	"reflect"
)

type DefaultHealthEventManager struct {
	evtChan                chan interface{}
	listenerEventTypeCache map[reflect.Type]Readiness
}

func NewDefaultHealthEventManager() *DefaultHealthEventManager {
	return &DefaultHealthEventManager{
		evtChan:                make(chan interface{}),
		listenerEventTypeCache: make(map[reflect.Type]Readiness),
	}
}

func (h *DefaultHealthEventManager) Dispatch(event interface{}) {
	select {
	case h.evtChan <- event:
	default:
	}
}

func (h *DefaultHealthEventManager) AddListener(listener Readiness) error {
	for _, evtType := range listener.EventTypes() {
		evtTypeVal := reflect.TypeOf(evtType)
		if _, ok := h.listenerEventTypeCache[evtTypeVal]; ok {
			return errors.New("event type already registered")
		}
		h.listenerEventTypeCache[reflect.TypeOf(evtType)] = listener
	}
	return nil
}

func (h *DefaultHealthEventManager) Watch(ctx context.Context) {
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
}
