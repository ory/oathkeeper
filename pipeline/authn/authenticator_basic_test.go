// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authn_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/internal"
	"github.com/ory/oathkeeper/pipeline/authn"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticatorBasic(t *testing.T) {
	t.Parallel()
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	session := new(authn.AuthenticationSession)

	a, err := reg.PipelineAuthenticator("basic")
	require.NoError(t, err)
	assert.Equal(t, "basic", a.GetID())

	t.Run("method=authenticate/case=empty auth", func(t *testing.T) {
		err := a.Authenticate(
			&http.Request{Header: http.Header{}},
			session,
			json.RawMessage(`{"credentials":"bc842c31a9e54efe320d30d948be61291f3ceee4766e36ab25fa65243cd76e0e"}`),
			nil)
			require.Error(t, err)
			assert.EqualError(t, err, helper.ErrUnauthorized.Error())
	})

	t.Run("method=authenticate/case=incorrect base64 token", func(t *testing.T) {
		err := a.Authenticate(
			&http.Request{Header: http.Header{ "Authorization": {"Basic " + "123"} }},
			session,
			json.RawMessage(`{"credentials":"bc842c31a9e54efe320d30d948be61291f3ceee4766e36ab25fa65243cd76e0e"}`),
			nil)
			require.Error(t, err)
			assert.EqualError(t, err, helper.ErrUnauthorized.Error())
	})

	t.Run("method=authenticate/case=incorrect auth", func(t *testing.T) {
		err := a.Authenticate(
			&http.Request{Header: http.Header{ "Authorization": {"Basic " + "dXNlcjpwYXNz"} }},
			session,
			json.RawMessage(`{"credentials":"bc842c31a9e54efe320d30d948be61291f3ceee4766e36ab25fa65243cd76e0e"}`),
			nil)
			require.Error(t, err)
			assert.EqualError(t, err, helper.ErrUnauthorized.Error())
	})

	t.Run("method=authenticate/case=correct auth", func(t *testing.T) {
		err := a.Authenticate(
			&http.Request{Header: http.Header{ "Authorization": {"Basic " + "dXNlcm5hbWU6cGFzc3dvcmQ="} }},
			session,
			json.RawMessage(`{"credentials":"bc842c31a9e54efe320d30d948be61291f3ceee4766e36ab25fa65243cd76e0e"}`),
			nil)
			require.NoError(t, err)
	})
}
