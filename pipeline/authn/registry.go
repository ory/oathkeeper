// Copyright © 2022 Ory Corp

package authn

type Registry interface {
	AvailablePipelineAuthenticators() []string
	PipelineAuthenticator(string) (Authenticator, error)
}
