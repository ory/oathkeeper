package errors_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/pipeline/errors"
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
			header      http.Header
			config      string
			expectError error
			givenError  error
			assert      func(t *testing.T, recorder *httptest.ResponseRecorder)
		}{
			{
				d:           "should return not responsible error because request is not application/json",
				header:      http.Header{"Accept": {"application/xml"}},
				expectError: errors.ErrHandlerNotResponsible,
				givenError:  &herodot.ErrNotFound,
				config:      `{"when":[{"request":{"header":{"accept":["application/json"]}}}]}`,
			},
			{
				d:           "should return not responsible error because the error types do not match",
				expectError: errors.ErrHandlerNotResponsible,
				givenError:  &herodot.ErrNotFound,
				config:      `{"when":[{"error":["unauthorized"]}]}`,
			},
			{
				d:          "should write to the request",
				givenError: &herodot.ErrNotFound,
				config:     `{"when":[{"error":["not_found"]}]}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					body := rw.Body.String()
					assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
					assert.Empty(t, gjson.Get(body, "error.reason").String())
					assert.Equal(t, int64(404), gjson.Get(body, "error.code").Int())
				},
			},
			{
				d:          "should write to the request handler and omit debug info because verbose is false",
				header:     http.Header{"Accept": {"application/json"}},
				givenError: herodot.ErrNotFound.WithReasonf("this should not show up in the response"),
				config:     `{"when":[{"request":{"header":{"accept":["application/json"]}}}]}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					body := rw.Body.String()
					assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
					assert.Empty(t, gjson.Get(body, "error.reason").String())
					assert.Equal(t, int64(404), gjson.Get(body, "error.code").Int())
				},
			},
			{
				d:          "should write to the request handler and include verbose error details",
				header:     http.Header{"Accept": {"application/json"}},
				givenError: herodot.ErrNotFound.WithReasonf("this must show up in the error details"),
				config:     `{"verbose": true, "when":[{"request":{"header":{"accept":["application/json"]}}}]}`,
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
				r.Header = tc.header
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
