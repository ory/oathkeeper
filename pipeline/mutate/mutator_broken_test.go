// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package mutate_test

import (
	"testing"

	"github.com/ory/oathkeeper/pipeline/mutate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialsIssuerBroken(t *testing.T) {
	t.Parallel()
	a := mutate.NewMutatorBroken(false)
	assert.Equal(t, "broken", a.GetID())

	err := a.Mutate(nil, nil, nil, nil)
	require.Error(t, err)

	t.Run("method=new/case=should not be declared in registry", func(t *testing.T) {
		err := a.Mutate(nil, nil, nil, nil)
		require.Error(t, err)
	})

	t.Run("method=validate", func(t *testing.T) {
		require.Error(t, mutate.NewMutatorBroken(false).Validate(nil))
		require.NoError(t, mutate.NewMutatorBroken(true).Validate(nil))
	})
}
