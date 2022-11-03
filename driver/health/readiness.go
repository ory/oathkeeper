// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
