// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package mutate

import (
	"github.com/ory/oathkeeper/credentials"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"

	"github.com/ory/oathkeeper/driver/configuration"
)

type Registry interface {
	AvailablePipelineMutators() []string
	PipelineMutator(string) (Mutator, error)
}

type dependencies interface {
	logrusx.Provider
	otelx.Provider
	configuration.Provider
	credentials.SignerRegistry
}
