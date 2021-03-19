package health

import (
	"context"
	"reflect"
)

type DefaultHealthEventManager struct {
	evtChan                 chan interface{}
	listeners               []Readiness
	listenerEventTypesCache map[Readiness][]reflect.Type
}

func NewDefaultHealthEventManager() *DefaultHealthEventManager {
	return &DefaultHealthEventManager{
		evtChan:                 make(chan interface{}),
		listeners:               []Readiness{},
		listenerEventTypesCache: make(map[Readiness][]reflect.Type),
	}
}

func (h *DefaultHealthEventManager) Dispatch(event interface{}) {
	select {
	case h.evtChan <- event:
	default:
	}
}

func (h *DefaultHealthEventManager) AddListener(listener Readiness) {
	h.listeners = append(h.listeners, listener)
	var typesCache []reflect.Type
	for _, evtType := range listener.EventTypes() {
		typesCache = append(typesCache, reflect.TypeOf(evtType))
	}
	h.listenerEventTypesCache[listener] = typesCache
}

func (h *DefaultHealthEventManager) internalDispatcher(evt interface{}, sourceEvtType reflect.Type) {
	for _, listener := range h.listeners {
		if evtTypesCache, ok := h.listenerEventTypesCache[listener]; ok {
			for _, evtType := range evtTypesCache {
				if sourceEvtType == evtType {
					listener.EventsReceiver(evt)
					return
				}
			}
		}
	}
}

func (h *DefaultHealthEventManager) Watch(ctx context.Context) {
	for {
		var evt interface{}
		select {
		case evt = <-h.evtChan:
		case <-ctx.Done():
			return
		}
		sourceEvtType := reflect.TypeOf(evt)
		h.internalDispatcher(evt, sourceEvtType)
	}
}
