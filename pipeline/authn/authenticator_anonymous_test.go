// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/pipeline/authn"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticatorAnonymous(t *testing.T) {
	t.Parallel()
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	session := new(authn.AuthenticationSession)

	a, err := reg.PipelineAuthenticator("anonymous")
	require.NoError(t, err)
	assert.Equal(t, "anonymous", a.GetID())

	t.Run("method=authenticate/case=is anonymous user", func(t *testing.T) {
		err := a.Authenticate(
			&http.Request{Header: http.Header{}},
			session,
			json.RawMessage(`{"subject":"anon"}`),
			nil)
		require.NoError(t, err)
		assert.Equal(t, "anon", session.Subject)
	})

	t.Run("method=authenticate/case=has credentials", func(t *testing.T) {
		err := a.Authenticate(
			&http.Request{Header: http.Header{"Authorization": {"foo"}}},
			session,
			json.RawMessage(`{"subject":"anon"}`),
			nil)
		require.Error(t, err)
	})

	t.Run("method=validate", func(t *testing.T) {
		conf.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, true)
		require.NoError(t, a.Validate(json.RawMessage(`{"subject":"foo"}`)))

		conf.SetForTest(t, configuration.AuthenticatorAnonymousIsEnabled, false)
		require.Error(t, a.Validate(json.RawMessage(`{"subject":"foo"}`)))
	})
}
