// Copyright © 2022 Ory Corp

package authz

type Registry interface {
	AvailablePipelineAuthorizers() []string
	PipelineAuthorizer(string) (Authorizer, error)
}
