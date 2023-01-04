// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
				d:          "redirect with 302 should contain a return_to param - absolute (HTTP) ",
				givenError: &herodot.ErrNotFound,
				config:     `{"to":"http://test/signin","return_to_query_param":"return_to"}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 302, rw.Code)
					location, err := url.Parse(rw.Header().Get("Location"))
					require.NoError(t, err)
					assert.Equal(t, "http://test/signin?return_to=%2Ftest", rw.Header().Get("Location"))
					assert.Equal(t, "/test", location.Query().Get("return_to"))
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
				d:          "redirect with 302 should contain a return_to param - absolute (HTTPS) ",
				givenError: &herodot.ErrNotFound,
				config:     `{"to":"https://test/signin","return_to_query_param":"return_to"}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 302, rw.Code)
					location, err := url.Parse(rw.Header().Get("Location"))
					require.NoError(t, err)
					assert.Equal(t, "https://test/signin?return_to=%2Ftest", rw.Header().Get("Location"))
					assert.Equal(t, "/test", location.Query().Get("return_to"))
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
				d:          "redirect with 302 should contain a return_to param - relative ",
				givenError: &herodot.ErrNotFound,
				config:     `{"to":"/test/signin","return_to_query_param":"return_to"}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 302, rw.Code)
					location, err := url.Parse(rw.Header().Get("Location"))
					require.NoError(t, err)
					assert.Equal(t, "/test/signin?return_to=%2Ftest", rw.Header().Get("Location"))
					assert.Equal(t, "/test", location.Query().Get("return_to"))
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
				d:          "redirect with 301 should contain a return_to param - absolute (HTTP) ",
				givenError: &herodot.ErrNotFound,
				config:     `{"to":"http://test/signin","return_to_query_param":"return_to","code":301}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 301, rw.Code)
					location, err := url.Parse(rw.Header().Get("Location"))
					require.NoError(t, err)
					assert.Equal(t, "http://test/signin?return_to=%2Ftest", rw.Header().Get("Location"))
					assert.Equal(t, "/test", location.Query().Get("return_to"))
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
				d:          "redirect with 301 should contain a return_to param - absolute (HTTPS) ",
				givenError: &herodot.ErrNotFound,
				config:     `{"to":"https://test/signin","return_to_query_param":"return_to","code":301}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 301, rw.Code)
					location, err := url.Parse(rw.Header().Get("Location"))
					require.NoError(t, err)
					assert.Equal(t, "https://test/signin?return_to=%2Ftest", rw.Header().Get("Location"))
					assert.Equal(t, "/test", location.Query().Get("return_to"))
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
				d:          "redirect with 301 should contain a return_to param - relative ",
				givenError: &herodot.ErrNotFound,
				config:     `{"to":"/test/signin","return_to_query_param":"return_to","code":301}`,
				assert: func(t *testing.T, rw *httptest.ResponseRecorder) {
					assert.Equal(t, 301, rw.Code)
					location, err := url.Parse(rw.Header().Get("Location"))
					require.NoError(t, err)
					assert.Equal(t, "/test/signin?return_to=%2Ftest", rw.Header().Get("Location"))
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

func TestErrorReturnToRedirectURLHeaderUsage(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	defaultUrl := &url.URL{Scheme: "http", Host: "ory.sh", Path: "/foo"}
	defaultTransform := func(req *http.Request) {}
	config := `{"to":"http://test/test","return_to_query_param":"return_to"}`

	a, err := reg.PipelineErrorHandler("redirect")
	require.NoError(t, err)
	assert.Equal(t, "redirect", a.GetID())

	for _, tc := range []struct {
		name        string
		expectedUrl *url.URL
		transform   func(req *http.Request)
	}{
		{
			name:        "all arguments are taken from the url and request method",
			expectedUrl: defaultUrl,
			transform:   defaultTransform,
		},
		{
			name:        "all arguments are taken from the headers",
			expectedUrl: &url.URL{Scheme: "https", Host: "test.dev", Path: "/bar"},
			transform: func(req *http.Request) {
				req.Header.Add("X-Forwarded-Proto", "https")
				req.Header.Add("X-Forwarded-Host", "test.dev")
				req.Header.Add("X-Forwarded-Uri", "/bar")
			},
		},
		{
			name:        "only scheme is taken from the headers",
			expectedUrl: &url.URL{Scheme: "https", Host: defaultUrl.Host, Path: defaultUrl.Path},
			transform: func(req *http.Request) {
				req.Header.Add("X-Forwarded-Proto", "https")
			},
		},
		{
			name:        "only host is taken from the headers",
			expectedUrl: &url.URL{Scheme: defaultUrl.Scheme, Host: "test.dev", Path: defaultUrl.Path},
			transform: func(req *http.Request) {
				req.Header.Add("X-Forwarded-Host", "test.dev")
			},
		},
		{
			name:        "only path is taken from the headers",
			expectedUrl: &url.URL{Scheme: defaultUrl.Scheme, Host: defaultUrl.Host, Path: "/bar"},
			transform: func(req *http.Request) {
				req.Header.Add("X-Forwarded-Uri", "/bar")
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", defaultUrl.String(), nil)
			tc.transform(r)

			err = a.Handle(w, r, json.RawMessage(config), nil, nil)
			assert.NoError(t, err)

			loc := w.Header().Get("Location")
			assert.NotEmpty(t, loc)

			locUrl, err := url.Parse(loc)
			assert.NoError(t, err)

			returnTo := locUrl.Query().Get("return_to")
			assert.NotEmpty(t, returnTo)

			returnToUrl, err := url.Parse(returnTo)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedUrl, returnToUrl)
		})
	}
}
