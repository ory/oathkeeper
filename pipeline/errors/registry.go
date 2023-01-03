// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package errors

type (
	Handlers []Handler
	Registry interface {
		AvailablePipelineErrorHandlers() Handlers
		PipelineErrorHandler(id string) (Handler, error)
	}
)

func (h Handlers) IDs() (res []string) {
	for _, hh := range h {
		res = append(res, hh.GetID())
	}
	return res
}
