/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package mutate_test

import (
	"testing"

	"github.com/ory/oathkeeper/pipeline/mutate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialsIssuerBroken(t *testing.T) {
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
