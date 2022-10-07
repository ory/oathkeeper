// Copyright © 2022 Ory Corp

package mutate

type Registry interface {
	AvailablePipelineMutators() []string
	PipelineMutator(string) (Mutator, error)
}
