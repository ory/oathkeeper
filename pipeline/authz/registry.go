package authz

type Registry interface {
	AvailablePipelineAuthorizers() []string
	PipelineAuthorizer(string) (Authorizer, error)
}
