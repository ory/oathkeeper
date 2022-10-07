// Copyright © 2022 Ory Corp

package health

type (
	ReadinessProbe interface {
		ID() string
		Validate() error

		EventTypes() []ReadinessProbeEvent
		EventsReceiver(event ReadinessProbeEvent)
	}

	ReadinessProbeEvent interface {
		ReadinessProbeListenerID() string
	}
)
