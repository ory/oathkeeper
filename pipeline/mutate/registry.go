package mutate

type Registry interface {
	AvailablePipelineMutators() []string
	PipelineMutator(string) (Mutator, error)
}
