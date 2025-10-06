// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
)

func TestRegistryMemoryAvailablePipelineAuthorizers(t *testing.T) {
	c, err := configuration.NewKoanfProvider(context.Background(), nil, logrusx.New("", ""))
	require.NoError(t, err)
	r := NewRegistry(c)
	got := r.AvailablePipelineAuthorizers()
	assert.ElementsMatch(t, got, []string{"allow", "deny", "keto_engine_acp_ory", "remote", "remote_json"})
}

func TestRegistryMemoryPipelineAuthorizer(t *testing.T) {
	tests := []struct {
		id      string
		wantErr bool
	}{
		{id: "allow"},
		{id: "deny"},
		{id: "keto_engine_acp_ory"},
		{id: "remote"},
		{id: "remote_json"},
		{id: "unregistered", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			c, err := configuration.NewKoanfProvider(context.Background(), nil, logrusx.New("", ""))
			require.NoError(t, err)
			r := NewRegistry(c)
			a, err := r.PipelineAuthorizer(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("PipelineAuthorizer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if a != nil && a.GetID() != tt.id {
				t.Errorf("PipelineAuthorizer() got = %v, want %v", a.GetID(), tt.id)
			}
		})
	}
}

func TestRegistryLoggerRedactsConfiguredHeaders(t *testing.T) {
	t.Run("case=logger redacts custom headers from configuration", func(t *testing.T) {
		c, err := configuration.NewKoanfProvider(
			context.Background(),
			nil,
			logrusx.New("", ""),
			configx.SkipValidation(),
			configx.WithValues(map[string]interface{}{
				"log.redact_headers": []string{"x-custom-authorization", "x-api-key"},
			}),
		)
		require.NoError(t, err)

		r := NewRegistry(c)
		logger := r.Logger()

		// Create headers with custom auth header
		headers := http.Header{}
		headers.Set("Authorization", "Bearer default-token")
		headers.Set("X-Custom-Authorization", "Bearer custom-token")
		headers.Set("X-API-Key", "secret-key")
		headers.Set("X-Request-ID", "request-123")

		// Get redacted headers
		redacted := logger.HTTPHeadersRedacted(headers)

		// Default sensitive headers should be redacted
		assert.Contains(t, redacted["authorization"], "Value is sensitive")

		// Custom configured headers should be redacted
		assert.Contains(t, redacted["x-custom-authorization"], "Value is sensitive")
		assert.Contains(t, redacted["x-api-key"], "Value is sensitive")

		// Non-sensitive headers should not be redacted
		assert.Equal(t, "request-123", redacted["x-request-id"])
	})

	t.Run("case=logger without custom configuration only redacts default headers", func(t *testing.T) {
		c, err := configuration.NewKoanfProvider(
			context.Background(),
			nil,
			logrusx.New("", ""),
			configx.SkipValidation(),
		)
		require.NoError(t, err)

		r := NewRegistry(c)
		logger := r.Logger()

		// Create headers with custom auth header
		headers := http.Header{}
		headers.Set("Authorization", "Bearer default-token")
		headers.Set("X-Custom-Authorization", "Bearer custom-token")
		headers.Set("X-API-Key", "secret-key")

		// Get redacted headers
		redacted := logger.HTTPHeadersRedacted(headers)

		// Default sensitive headers should be redacted
		assert.Contains(t, redacted["authorization"], "Value is sensitive")

		// Custom headers should NOT be redacted (no configuration)
		assert.Equal(t, "Bearer custom-token", redacted["x-custom-authorization"])
		assert.Equal(t, "secret-key", redacted["x-api-key"])
	})
}
