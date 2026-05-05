// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"testing"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/ory/x/logrusx"
)

type TestLoggerProvider struct {
	T *testing.T
}

func (lp *TestLoggerProvider) Logger() *logrusx.Logger {
	return logrusx.NewT(lp.T)
}

func (lp *TestLoggerProvider) Tracer() trace.Tracer {
	return noop.NewTracerProvider().Tracer(lp.T.Name())
}
