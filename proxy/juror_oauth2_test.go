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

package proxy

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/sirupsen/logrus"
	"github.com/ory/oathkeeper/rule"
	"net/url"
	"net/http"
	"fmt"
	"github.com/stretchr/testify/require"
)

func TestOAuth2Juror(t *testing.T) {
	t.Run("suite=regular", func(t *testing.T) {
		j := &JurorOAuth2Introspection{L: logrus.New()}
		assert.Equal(t, "oauth2_introspection", j.GetID())

		rl := &rule.Rule{ID: "1234", Mode: "foo"}
		for k, tc := range []struct {
			token         string
			expectErr     bool
			expectSession *Session
		}{
			{
				token: "foo",

			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				r := &http.Request{Header: http.Header{"Authorization": {"Bearer " + tc.token}}}
				s, err := j.Try(r, rl, new(url.URL))

				if tc.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.EqualValues(t, tc.expectSession, s)
				}
			})
		}
	})

	t.Run("suite=anonymous", func(t *testing.T) {
		ja := &JurorOAuth2Introspection{
			AllowAnonymous: true,
			L:              logrus.New(),
		}
		assert.Equal(t, "oauth2_introspection_anonymous", ja.GetID())
	})
}
