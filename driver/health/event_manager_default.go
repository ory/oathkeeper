// Copyright 2021 Ory GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package health

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ory/x/healthx"
)

type DefaultHealthEventManager struct {
	evtChan                chan ReadinessProbeEvent
	listeners              []ReadinessProbe
	listenerEventTypeCache map[string]ReadinessProbe
}

func NewDefaultHealthEventManager(listeners ...ReadinessProbe) (*DefaultHealthEventManager, error) {
	var listenerEventTypeCache = make(map[string]ReadinessProbe)
	for _, listener := range listeners {
		for _, events := range listener.EventTypes() {
			if _, ok := listenerEventTypeCache[events.ReadinessProbeListenerID()]; ok {
				return nil, errors.WithStack(ErrEventTypeAlreadyRegistered)
			}
			listenerEventTypeCache[events.ReadinessProbeListenerID()] = listener
		}
	}
	return &DefaultHealthEventManager{
		evtChan:                make(chan ReadinessProbeEvent),
		listeners:              listeners,
		listenerEventTypeCache: listenerEventTypeCache,
	}, nil
}

func (h *DefaultHealthEventManager) Dispatch(event ReadinessProbeEvent) {
	if event == nil {
		return
	}
	go func() {
		h.evtChan <- event
	}()
}

func (h *DefaultHealthEventManager) Watch(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case evt, ok := <-h.evtChan:
				if !ok {
					return
				}
				if listener, ok := h.listenerEventTypeCache[evt.ReadinessProbeListenerID()]; ok {
					listener.EventsReceiver(evt)
				}
			}
		}
	}()
}

func (h *DefaultHealthEventManager) HealthxReadyCheckers() healthx.ReadyCheckers {
	var checkers = make(healthx.ReadyCheckers)
	for _, listener := range h.listeners {
		checkers[listener.ID()] = listener.Validate
	}
	return checkers
}
