// Copyright © 2022 Ory Corp

package pipeline

import "github.com/pkg/errors"

var (
	ErrPipelineHandlerNotFound = errors.New("requested pipeline handler does not exist")
)
