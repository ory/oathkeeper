// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package errors_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"

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
				givenError: &herodot.ErrNotFound,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					body := rw.Body.String()
					assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
					assert.Empty(t, gjson.Get(body, "error.reason").String())
					assert.Equal(t, int64(404), gjson.Get(body, "error.code").Int())
				},
			},
			{
				d:          "should write to the request handler and omit debug info because verbose is false",
				givenError: herodot.ErrNotFound.WithReasonf("this should not show up in the response"),
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					body := rw.Body.String()
					assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
					assert.Empty(t, gjson.Get(body, "error.reason").String())
					assert.Equal(t, int64(404), gjson.Get(body, "error.code").Int())
				},
			},
			{
				d:          "should write to the request handler and include verbose error details",
				givenError: herodot.ErrNotFound.WithReasonf("this must show up in the error details"),
				config:     `{"verbose": true}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					body := rw.Body.String()
					assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
					assert.Equal(t, "this must show up in the error details", gjson.Get(body, "error.reason").String())
					assert.Equal(t, int64(404), gjson.Get(body, "error.code").Int())
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
