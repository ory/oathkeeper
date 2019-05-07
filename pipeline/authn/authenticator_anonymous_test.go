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
	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	. "github.com/ory/oathkeeper/pipeline/authn"
	"github.com/spf13/viper"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticatorAnonymous(t *testing.T) {
	conf := internal.NewConfigurationWithDefaults()
	viper.Set(configuration.ViperKeyAuthenticatorAnonymousIdentifier, "anon")

	a := NewAuthenticatorAnonymous(conf)
	assert.Equal(t, "anonymous", a.GetID())

	t.Run("case=is anonymous user", func(t *testing.T) {
		session, err := a.Authenticate(&http.Request{Header: http.Header{}}, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "anon", session.Subject)
	})

	t.Run("case=has credentials", func(t *testing.T) {
		_, err := a.Authenticate(&http.Request{Header: http.Header{"Authorization": {"foo"}}}, nil, nil)
		require.Error(t, err)
	})

	t.Run("case=validate enabled/disabled", func(t *testing.T) {
		viper.Set(configuration.ViperKeyAuthenticatorAnonymousIsEnabled, true)
		require.NoError(t, a.Validate())

		viper.Set(configuration.ViperKeyAuthenticatorAnonymousIsEnabled, false)
		require.Error(t, a.Validate())
	})
}
