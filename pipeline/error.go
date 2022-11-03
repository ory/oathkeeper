// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package pipeline

import "github.com/pkg/errors"

var (
	ErrPipelineHandlerNotFound = errors.New("requested pipeline handler does not exist")
)
