// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn_test

import (
	"testing"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticatorNoop(t *testing.T) {
	t.Parallel()
	reg := internal.NewRegistry(t)

	a, err := reg.PipelineAuthenticator("noop")
	require.NoError(t, err)
	assert.Equal(t, "noop", a.GetID())

	t.Run("method=authenticate", func(t *testing.T) {
		err := a.Authenticate(nil, nil, nil, nil)
		require.NoError(t, err)
	})

	t.Run("method=validate", func(t *testing.T) {
		reg.Config().SetForTest(t, configuration.AuthenticatorNoopIsEnabled, true)
		require.NoError(t, a.Validate(nil))

		reg.Config().SetForTest(t, configuration.AuthenticatorNoopIsEnabled, false)
		require.Error(t, a.Validate(nil))
	})
}
