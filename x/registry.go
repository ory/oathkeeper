// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"go.opentelemetry.io/otel/trace"

	"github.com/ory/x/logrusx"
)

type TestLoggerProvider struct{}

func (lp *TestLoggerProvider) Logger() *logrusx.Logger {
	return logrusx.New("", "")
}

func (lp *TestLoggerProvider) Tracer() trace.Tracer {
	return nil
}
