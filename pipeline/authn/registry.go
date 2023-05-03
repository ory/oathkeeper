// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn

type Registry interface {
	AvailablePipelineAuthenticators() []string
	PipelineAuthenticator(string) (Authenticator, error)
}
