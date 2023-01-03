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

func TestAuthorizerAllow(t *testing.T) {
	t.Parallel()
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineAuthorizer("allow")
	require.NoError(t, err)
	assert.Equal(t, "allow", a.GetID())

	t.Run("method=authorize/case=passes always", func(t *testing.T) {
		require.NoError(t, a.Authorize(nil, nil, nil, nil))
	})

	t.Run("method=validate", func(t *testing.T) {
		conf.SetForTest(t, configuration.AuthorizerAllowIsEnabled, true)
		require.NoError(t, a.Validate(nil))

		conf.SetForTest(t, configuration.AuthorizerAllowIsEnabled, false)
		require.Error(t, a.Validate(nil))
	})
}
