// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
