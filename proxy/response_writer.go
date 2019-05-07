/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

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
