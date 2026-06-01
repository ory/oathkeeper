// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz

import (
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"

	"github.com/ory/oathkeeper/driver/configuration"
)

type Registry interface {
	AvailablePipelineAuthorizers() []string
	PipelineAuthorizer(string) (Authorizer, error)
}

type dependencies interface {
	logrusx.Provider
	otelx.Provider
	configuration.Provider
}
