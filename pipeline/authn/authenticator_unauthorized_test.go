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

package authn_test

import (
	"net/http"
	"testing"

	"github.com/ory/viper"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticatorBroken(t *testing.T) {
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
		viper.Set(configuration.ViperKeyAuthenticatorUnauthorizedIsEnabled, true)
		require.NoError(t, a.Validate(nil))

		viper.Reset()
		viper.Set(configuration.ViperKeyAuthenticatorUnauthorizedIsEnabled, false)
		require.Error(t, a.Validate(nil))
	})
}
