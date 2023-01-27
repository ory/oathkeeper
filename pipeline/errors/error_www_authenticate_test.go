// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package errors_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/internal"
)

func TestErrorWWWAuthenticate(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineErrorHandler("www_authenticate")
	require.NoError(t, err)
	assert.Equal(t, "www_authenticate", a.GetID())

	t.Run("method=handle", func(t *testing.T) {
		for k, tc := range []struct {
			d           string
			header      http.Header
			config      string
			expectError error
			givenError  error
			assert      func(t *testing.T, recorder *httptest.ResponseRecorder)
		}{
			{
				d:          "should respond with a 401 realm message",
				givenError: &herodot.ErrNotFound,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 401, rw.Code)
					assert.Equal(t, "Basic realm=Please authenticate.", rw.Header().Get("WWW-Authenticate"))
				},
			},
			{
				d:          "should respond with a 401 realm message and a custom message",
				config:     `{"realm": "foobar"}`,
				givenError: &herodot.ErrNotFound,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 401, rw.Code)
					assert.Equal(t, "Basic realm=foobar", rw.Header().Get("WWW-Authenticate"))
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
