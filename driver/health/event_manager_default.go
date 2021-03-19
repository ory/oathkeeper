package health

import (
	"context"
	"reflect"
)

type DefaultHealthEventManager struct {
	evtChan   chan interface{}
	listeners []Readiness
}

func NewDefaultHealthEventManager() *DefaultHealthEventManager {
	return &DefaultHealthEventManager{
		evtChan: make(chan interface{}),
		listeners: []Readiness{},
	}
}

func (h *DefaultHealthEventManager) Dispatch(event interface{}) {
	h.evtChan <- event
}

func (h *DefaultHealthEventManager) AddListener(listener Readiness) {
	h.listeners = append(h.listeners, listener)
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
		for _, listener := range h.listeners {
			for _, evtType := range listener.EventTypes() {
				if sourceEvtType == reflect.TypeOf(evtType) {
					listener.EventsReceiver(evt)
				}
			}
		}
	}
}
