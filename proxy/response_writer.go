// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"bytes"
	"net/http"
)

type simpleResponseWriter struct {
	header http.Header
	code   int
	buffer *bytes.Buffer
}

func NewSimpleResponseWriter() *simpleResponseWriter {
	return &simpleResponseWriter{
		header: http.Header{},
		buffer: new(bytes.Buffer),
	}
}

func (r *simpleResponseWriter) Header() http.Header {
	return r.header
}

func (r *simpleResponseWriter) Write(b []byte) (int, error) {
	return r.buffer.Write(b)
}

func (r *simpleResponseWriter) WriteHeader(statusCode int) {
	r.code = statusCode
}
