// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn_test

import (
	"net/http"
	"testing"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticatorBroken(t *testing.T) {
	t.Parallel()
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineAuthenticator("unauthorized")
	require.NoError(t, err)
	assert.Equal(t, "unauthorized", a.GetID())

	t.Run("method=authenticate", func(t *testing.T) {
		err := a.Authenticate(&http.Request{Header: http.Header{}}, nil, nil, nil)
		require.Error(t, err)
	})

	t.Run("method=validate", func(t *testing.T) {
		conf.SetForTest(t, configuration.AuthenticatorUnauthorizedIsEnabled, true)
		require.NoError(t, a.Validate(nil))

		conf.SetForTest(t, configuration.AuthenticatorUnauthorizedIsEnabled, false)
		require.Error(t, a.Validate(nil))
	})
}
