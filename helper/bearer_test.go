// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package helper_test

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/oathkeeper/helper"
)

const (
	defaultHeaderName = "Authorization"
)

func TestBearerTokenFromRequest(t *testing.T) {
	t.Run("case=token should be received from default header if custom location is not set and token is present", func(t *testing.T) {
		expectedToken := "token"
		request := &http.Request{Header: http.Header{defaultHeaderName: {"bearer " + expectedToken}}}
		token := helper.BearerTokenFromRequest(request, nil)
		assert.Equal(t, expectedToken, token)
	})
	t.Run("case=should return empty string if custom location is not set and token is not present in default header", func(t *testing.T) {
		request := &http.Request{}
		token := helper.BearerTokenFromRequest(request, nil)
		assert.Empty(t, token)
	})
	t.Run("case=should return empty string if custom location is set to header and token is not present in that header", func(t *testing.T) {
		customHeaderName := "Custom-Auth-Header"
		request := &http.Request{Header: http.Header{defaultHeaderName: {"bearer token"}}}
		tokenLocation := helper.BearerTokenLocation{Header: &customHeaderName}
		token := helper.BearerTokenFromRequest(request, &tokenLocation)
		assert.Empty(t, token)
	})
	t.Run("case=should return empty string if custom location is set to query parameter and token is not present in that query parameter", func(t *testing.T) {
		customQueryParameterName := "Custom-Auth"
		request := &http.Request{Header: http.Header{defaultHeaderName: {"bearer token"}}}
		tokenLocation := helper.BearerTokenLocation{QueryParameter: &customQueryParameterName}
		token := helper.BearerTokenFromRequest(request, &tokenLocation)
		assert.Empty(t, token)
	})
	t.Run("case=token should be received from custom header if custom location is set to header and token is present", func(t *testing.T) {
		expectedToken := "token"
		customHeaderName := "Custom-Auth-Header"
		request := &http.Request{Header: http.Header{customHeaderName: {expectedToken}}}
		tokenLocation := helper.BearerTokenLocation{Header: &customHeaderName}
		token := helper.BearerTokenFromRequest(request, &tokenLocation)
		assert.Equal(t, expectedToken, token)
	})
	t.Run("case=token should be received from custom header if custom location is set to query parameter and token is present", func(t *testing.T) {
		expectedToken := "token"
		customQueryParameterName := "Custom-Auth"
		request := &http.Request{
			URL: &url.URL{RawQuery: customQueryParameterName + "=" + expectedToken},
		}
		tokenLocation := helper.BearerTokenLocation{QueryParameter: &customQueryParameterName}
		token := helper.BearerTokenFromRequest(request, &tokenLocation)
		assert.Equal(t, expectedToken, token)
	})
	t.Run("case=token should be received from default header if custom token location is set, but neither Header nor Query Param is configured", func(t *testing.T) {
		expectedToken := "token"
		request := &http.Request{Header: http.Header{defaultHeaderName: {"bearer " + expectedToken}}}
		tokenLocation := helper.BearerTokenLocation{}
		token := helper.BearerTokenFromRequest(request, &tokenLocation)
		assert.Equal(t, expectedToken, token)
	})
	t.Run("case=token should be received from default header if custom location is set to default location and token is present", func(t *testing.T) {
		expectedToken := "token"
		customHeaderName := "Authorization"
		request := &http.Request{Header: http.Header{customHeaderName: {"bearer " + expectedToken}}}
		tokenLocation := helper.BearerTokenLocation{Header: &customHeaderName}
		token := helper.BearerTokenFromRequest(request, &tokenLocation)
		assert.Equal(t, expectedToken, token)
	})

	t.Run("case=Should not consume body when token from query parameter and Content-type is 'application/x-www-form-urlencoded' ", func(t *testing.T) {
		expectedToken := "token"
		customQueryParameterName := "Custom-Auth"

		request := &http.Request{
			Method: http.MethodPost,
			URL: &url.URL{
				RawQuery: customQueryParameterName + "=" + expectedToken,
			},
			Body: io.NopCloser(strings.NewReader("body")),
			Header: http.Header{
				"Content-Type": {"application/x-www-form-urlencoded"},
			},
		}

		tokenLocation := helper.BearerTokenLocation{QueryParameter: &customQueryParameterName}
		token := helper.BearerTokenFromRequest(request, &tokenLocation)
		assert.Equal(t, expectedToken, token)

		b, err := io.ReadAll(request.Body)
		assert.NoError(t, err)
		assert.Equal(t, "body", string(b))
	})
}
