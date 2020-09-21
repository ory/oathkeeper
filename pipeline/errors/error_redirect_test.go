package errors_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"

	"github.com/ory/oathkeeper/internal"
)

func TestErrorRedirect(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineErrorHandler("redirect")
	require.NoError(t, err)
	assert.Equal(t, "redirect", a.GetID())

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
				d:          "should redirect with 302 - absolute (HTTP)",
				givenError: &herodot.ErrNotFound,
				config:     `{"to":"http://test/test"}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 302, rw.Code)
					assert.Equal(t, "http://test/test", rw.Header().Get("Location"))
				},
			},
			{
				d:          "should redirect with 302 - absolute (HTTPS)",
				givenError: &herodot.ErrNotFound,
				config:     `{"to":"https://test/test"}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 302, rw.Code)
					assert.Equal(t, "https://test/test", rw.Header().Get("Location"))
				},
			},
			{
				d:          "should redirect with 302 - relative",
				givenError: &herodot.ErrNotFound,
				config:     `{"to":"/test"}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 302, rw.Code)
					assert.Equal(t, "/test", rw.Header().Get("Location"))
				},
			},
			{
				d:          "should redirect with 301 - absolute (HTTP)",
				givenError: &herodot.ErrNotFound,
				config:     `{"to":"http://test/test","code":301}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 301, rw.Code)
					assert.Equal(t, "http://test/test", rw.Header().Get("Location"))
				},
			},
			{
				d:          "should redirect with 301 - absolute (HTTPS)",
				givenError: &herodot.ErrNotFound,
				config:     `{"to":"https://test/test","code":301}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 301, rw.Code)
					assert.Equal(t, "https://test/test", rw.Header().Get("Location"))
				},
			},
			{
				d:          "should redirect with 301 - relative",
				givenError: &herodot.ErrNotFound,
				config:     `{"to":"/test", "code":301}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 301, rw.Code)
					assert.Equal(t, "/test", rw.Header().Get("Location"))
				},
			},
			{
				d:          "should redirect with return_to param",
				givenError: &herodot.ErrNotFound,
				config:     `{"to":"http://test/signin","return_to_query_param":"return_to"}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 302, rw.Code)
					location, err := url.Parse(rw.Header().Get("Location"))
					require.NoError(t, err)
					assert.Equal(t, "/test", location.Query().Get("return_to"))
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
