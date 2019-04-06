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

package health

import (
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ory/oathkeeper/sdk/go/oathkeeper/client"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
)

func TestHealth(t *testing.T) {
	alive := errors.New("not alive")
	handler := &Handler{
		H:             herodot.NewJSONWriter(nil),
		VersionString: "test version",
		ReadyChecks: map[string]ReadyChecker{
			"test": func() error {
				return alive
			},
		},
	}
	router := httprouter.New()
	handler.SetRoutes(router)
	ts := httptest.NewServer(router)

	u, err := url.ParseRequestURI(ts.URL)
	require.NoError(t, err)
	cl := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:     u.Host,
		BasePath: u.Path,
		Schemes:  []string{u.Scheme},
	})

	aliveRes, err := cl.Health.IsInstanceAlive(nil)
	require.NoError(t, err)
	assert.EqualValues(t, "ok", aliveRes.Payload.Status)

	versionRes, err := cl.Version.GetVersion(nil)
	require.NoError(t, err)
	require.EqualValues(t, versionRes.Payload.Version, handler.VersionString)

	_, err = cl.Health.IsInstanceReady(nil)
	require.Error(t, err)

	alive = nil
	readyRes, err := cl.Health.IsInstanceReady(nil)
	require.NoError(t, err)
	assert.EqualValues(t, "ok", readyRes.Payload.Status)
}
