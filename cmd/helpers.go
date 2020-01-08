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

package cmd

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/ory/oathkeeper/internal/httpclient/client"
	"github.com/ory/x/cmdx"
	"github.com/ory/x/flagx"
)

func newClient(cmd *cobra.Command) *client.OryOathkeeper {
	endpoint := flagx.MustGetString(cmd, "endpoint")
	if endpoint == "" {
		cmdx.Fatalf("Please specify the endpoint url using the --endpoint flag, for more information use `oathkeeper help rules`")
	}

	u, err := url.ParseRequestURI(endpoint)
	cmdx.Must(err, `Unable to parse endpoint URL "%s": %s`, endpoint, err)

	return client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:     u.Host,
		BasePath: u.Path,
		Schemes:  []string{u.Scheme},
	})
}
