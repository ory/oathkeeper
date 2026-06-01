// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
)

func TestRegistryMemoryAvailablePipelineAuthorizers(t *testing.T) {
	l := logrusx.NewT(t)
	c, err := configuration.NewKoanfProvider(t.Context(), nil, l)
	require.NoError(t, err)
	r := NewRegistry(c, l)
	got := r.AvailablePipelineAuthorizers()
	assert.ElementsMatch(t, got, []string{"allow", "deny", "keto_engine_acp_ory", "remote", "remote_json"})
}

func TestRegistryMemoryPipelineAuthorizer(t *testing.T) {
	for _, tc := range []struct {
		id          string
		expectedErr error
	}{
		{id: "allow"},
		{id: "deny"},
		{id: "keto_engine_acp_ory"},
		{id: "remote"},
		{id: "remote_json"},
		{id: "unregistered", expectedErr: pipeline.ErrPipelineHandlerNotFound},
	} {
		t.Run(tc.id, func(t *testing.T) {
			l := logrusx.NewT(t)
			c, err := configuration.NewKoanfProvider(t.Context(), nil, l)
			require.NoError(t, err)
			r := NewRegistry(c, l)
			a, err := r.PipelineAuthorizer(tc.id)
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
				return
			}
			assert.Equal(t, tc.id, a.GetID())
		})
	}
}
