// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package mutate_test

import (
	"net/http"
	"testing"

	"github.com/ory/oathkeeper/pipeline/authn"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMutatorNoop(t *testing.T) {
	t.Parallel()
	conf := internal.NewConfigurationWithDefaults()
	reg := internal.NewRegistry(conf)

	a, err := reg.PipelineMutator("noop")
	require.NoError(t, err)
	assert.Equal(t, "noop", a.GetID())

	t.Run("method=mutate/case=passes always", func(t *testing.T) {
		r := &http.Request{Header: http.Header{"foo": {"foo"}}}
		s := &authn.AuthenticationSession{}
		err := a.Mutate(r, s, nil, nil)
		require.NoError(t, err)
		assert.EqualValues(t, r.Header, s.Header)
	})

	t.Run("method=mutate/case=ensure authentication session headers are kept", func(t *testing.T) {
		r := &http.Request{Header: http.Header{"foo": {"foo"}}}
		s := &authn.AuthenticationSession{Header: http.Header{"bar": {"bar"}}}
		combinedHeaders := http.Header{"foo": {"foo"}}
		combinedHeaders.Set("bar", "bar")
		err := a.Mutate(r, s, nil, nil)
		require.NoError(t, err)
		assert.EqualValues(t, r.Header, combinedHeaders)
	})

	t.Run("method=validate", func(t *testing.T) {
		conf.SetForTest(t, configuration.MutatorNoopIsEnabled, true)
		require.NoError(t, a.Validate(nil))

		conf.SetForTest(t, configuration.MutatorNoopIsEnabled, false)
		require.Error(t, a.Validate(nil))
	})
}
