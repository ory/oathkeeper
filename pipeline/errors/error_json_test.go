// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package errors_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/internal"
)

func TestErrorJSON(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineErrorHandler("json")
	require.NoError(t, err)
	assert.Equal(t, "json", a.GetID())

	t.Run("method=handle", func(t *testing.T) {
		for k, tc := range []struct {
			d           string
			config      string
			expectError error
			givenError  error
			assert      func(t *testing.T, recorder *httptest.ResponseRecorder)
		}{
			{
				d:          "should write to the request",
				givenError: herodot.ErrNotFound(),
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					body := rw.Body.String()
					assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
					assert.Empty(t, gjson.Get(body, "error.reason").String())
					assert.Equal(t, int64(404), gjson.Get(body, "error.code").Int())
				},
			},
			{
				d:          "should write to the request handler and omit debug info because verbose is false",
				givenError: herodot.ErrNotFound().WithReasonf("this should not show up in the response"),
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					body := rw.Body.String()
					assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
					assert.Empty(t, gjson.Get(body, "error.reason").String())
					assert.Equal(t, int64(404), gjson.Get(body, "error.code").Int())
				},
			},
			{
				d:          "should write to the request handler and include verbose error details",
				givenError: herodot.ErrNotFound().WithReasonf("this must show up in the error details"),
				config:     `{"verbose": true}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					body := rw.Body.String()
					assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
					assert.Equal(t, "this must show up in the error details", gjson.Get(body, "error.reason").String())
					assert.Equal(t, int64(404), gjson.Get(body, "error.code").Int())
				},
			},
			{
				d: "should propagate rate-limit headers from ErrWithHeaders in non-verbose mode",
				givenError: errors.WithStack(&helper.ErrWithHeaders{
					Err: helper.ErrTooManyRequests(),
					Headers: http.Header{
						"Retry-After":           []string{"60"},
						"X-RateLimit-Limit":     []string{"100"},
						"X-RateLimit-Remaining": []string{"0"},
						"X-RateLimit-Reset":     []string{"1234567890"},
					},
				}),
				config: `{"verbose": false}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					// Verify headers are written as HTTP response headers
					assert.Equal(t, "60", rw.Header().Get("Retry-After"))
					assert.Equal(t, "100", rw.Header().Get("X-RateLimit-Limit"))
					assert.Equal(t, "0", rw.Header().Get("X-RateLimit-Remaining"))
					assert.Equal(t, "1234567890", rw.Header().Get("X-RateLimit-Reset"))

					// Verify the error body is still 429
					body := rw.Body.String()
					assert.Equal(t, int64(429), gjson.Get(body, "error.code").Int())
					assert.Equal(t, "Too many requests", gjson.Get(body, "error.message").String())
				},
			},
			{
				d: "should propagate rate-limit headers from ErrWithHeaders in verbose mode",
				givenError: &helper.ErrWithHeaders{
					Err: helper.ErrTooManyRequests(),
					Headers: http.Header{
						"Retry-After": []string{"30"},
					},
				},
				config: `{"verbose": true}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					// Verify headers are written even in verbose mode
					assert.Equal(t, "30", rw.Header().Get("Retry-After"))

					// Verify verbose mode preserves the original error structure
					body := rw.Body.String()
					assert.Equal(t, int64(429), gjson.Get(body, "error.code").Int())
				},
			},
			{
				d: "should handle ErrWithHeaders with no headers gracefully",
				givenError: &helper.ErrWithHeaders{
					Err:     helper.ErrTooManyRequests(),
					Headers: http.Header{}, // Empty headers
				},
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					// Verify no rate-limit headers are written
					assert.Empty(t, rw.Header().Get("Retry-After"))
					assert.Empty(t, rw.Header().Get("X-RateLimit-Limit"))

					// Verify the error is still 429
					body := rw.Body.String()
					assert.Equal(t, int64(429), gjson.Get(body, "error.code").Int())
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
				w := httptest.NewRecorder()
				r := httptest.NewRequest("GET", "/test", nil)

				err := a.Handle(w, r, json.RawMessage(tc.config), nil, tc.givenError)
				if tc.expectError != nil {
					require.EqualError(t, err, tc.expectError.Error(), "%+v", err)
					return
				}

				require.NoError(t, err)
				if tc.assert != nil {
					tc.assert(t, w)
				}
			})
		}
	})
}
