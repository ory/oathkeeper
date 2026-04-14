// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package helper_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/oathkeeper/helper"
)

func TestErrWithHeaders(t *testing.T) {
	t.Run("Error() delegates to wrapped error", func(t *testing.T) {
		err := &helper.ErrWithHeaders{
			Err:     helper.ErrTooManyRequests(),
			Headers: http.Header{},
		}
		assert.Equal(t, "Too many requests", err.Error())
	})

	t.Run("Unwrap() returns wrapped error", func(t *testing.T) {
		err := &helper.ErrWithHeaders{
			Err:     helper.ErrTooManyRequests(),
			Headers: http.Header{},
		}
		assert.Equal(t, helper.ErrTooManyRequests(), err.Unwrap())
	})

	t.Run("errors.As finds ErrWithHeaders in chain", func(t *testing.T) {
		base := &helper.ErrWithHeaders{
			Err: helper.ErrTooManyRequests(),
			Headers: http.Header{
				"Retry-After": []string{"60"},
			},
		}

		var target *helper.ErrWithHeaders
		assert.True(t, errors.As(base, &target))
		assert.Equal(t, "60", target.Headers.Get("Retry-After"))
	})
}

func TestNewErrTooManyRequestsWithHeaders(t *testing.T) {
	t.Run("captures all rate-limit headers", func(t *testing.T) {
		resp := &http.Response{
			Header: http.Header{
				"Retry-After":           []string{"120"},
				"X-Ratelimit-Limit":     []string{"1000"},
				"X-Ratelimit-Remaining": []string{"0"},
				"X-Ratelimit-Reset":     []string{"1234567890"},
				"X-Ratelimit-Type":      []string{"user"},
				"Content-Type":          []string{"application/json"}, // Not in list
			},
		}

		err := helper.NewErrTooManyRequestsWithHeaders(resp)

		assert.Equal(t, "120", err.Headers.Get("Retry-After"))
		assert.Equal(t, "1000", err.Headers.Get("X-Ratelimit-Limit"))
		assert.Equal(t, "0", err.Headers.Get("X-Ratelimit-Remaining"))
		assert.Equal(t, "1234567890", err.Headers.Get("X-Ratelimit-Reset"))
		assert.Equal(t, "user", err.Headers.Get("X-Ratelimit-Type"))

		// Verify non-rate-limit headers are NOT captured
		assert.Empty(t, err.Headers.Get("Content-Type"))
	})

	t.Run("handles response with no rate-limit headers", func(t *testing.T) {
		resp := &http.Response{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		}

		err := helper.NewErrTooManyRequestsWithHeaders(resp)

		assert.Empty(t, err.Headers.Get("Retry-After"))
		assert.Empty(t, err.Headers.Get("X-RateLimit-Limit"))
		assert.NotNil(t, err.Headers) // Headers map exists but is empty
	})

	t.Run("wraps ErrTooManyRequests", func(t *testing.T) {
		resp := &http.Response{Header: http.Header{}}
		err := helper.NewErrTooManyRequestsWithHeaders(resp)

		assert.Equal(t, helper.ErrTooManyRequests(), err.Err)
		assert.Equal(t, "Too many requests", err.Error())
	})
}
