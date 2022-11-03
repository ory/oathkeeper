// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package mutate

type Registry interface {
	AvailablePipelineMutators() []string
	PipelineMutator(string) (Mutator, error)
}
