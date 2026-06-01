// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz_test

import (
	"testing"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizerDeny(t *testing.T) {
	t.Parallel()
	reg := internal.NewRegistry(t)

	a, err := reg.PipelineAuthorizer("deny")
	require.NoError(t, err)
	assert.Equal(t, "deny", a.GetID())

	t.Run("method=authorize/case=always returns denied", func(t *testing.T) {
		require.Error(t, a.Authorize(nil, nil, nil, nil))
	})

	t.Run("method=validate", func(t *testing.T) {
		reg.Config().SetForTest(t, configuration.AuthorizerDenyIsEnabled, true)
		require.NoError(t, a.Validate(nil))

		reg.Config().SetForTest(t, configuration.AuthorizerDenyIsEnabled, false)
		require.Error(t, a.Validate(nil))
	})
}
