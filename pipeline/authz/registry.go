// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz

type Registry interface {
	AvailablePipelineAuthorizers() []string
	PipelineAuthorizer(string) (Authorizer, error)
}
