package authn

type Registry interface {
	AvailablePipelineAuthenticators() []string
	PipelineAuthenticator(string) (Authenticator, error)
}
