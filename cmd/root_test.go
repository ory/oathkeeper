// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/ory/x/cmdx"

	"github.com/phayes/freeport"
)

var apiPort, proxyPort int

func init() {
	p, err := freeport.GetFreePorts(2)
	if err != nil {
		panic(err)
	}
	apiPort, proxyPort = p[0], p[1]

	os.Setenv("SERVE_API_PORT", fmt.Sprintf("%d", apiPort))
	os.Setenv("SERVE_PROXY_PORT", fmt.Sprintf("%d", proxyPort))
	os.Setenv("AUTHENTICATORS_NOOP_ENABLED", "1")
	os.Setenv("AUTHENTICATORS_ANONYMOUS_ENABLED", "true")
	os.Setenv("AUTHORIZERS_ALLOW_ENABLED", "true")
	os.Setenv("MUTATORS_NOOP_ENABLED", "true")
	os.Setenv("ACCESS_RULES_REPOSITORIES", "inline://"+base64.StdEncoding.EncodeToString([]byte(`[
  {
    "id": "test-rule-4",
    "upstream": {
      "preserve_host": true,
      "strip_path": "/api",
      "url": "https://mybackend.com/api"
    },
    "match": {
      "url": "myproxy.com/api",
      "methods": [
        "GET",
        "POST"
      ]
    },
    "authenticators": [
      {
        "handler": "noop"
      },
      {
        "handler": "anonymous"
      }
    ],
    "authorizer": {
      "handler": "allow"
    },
    "mutators": [
      {
        "handler": "noop"
      }
    ]
  }
]`)))
}

func ensureOpen(t *testing.T, port int) bool {
	res, err := http.Get(fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		t.Logf("Network error while polling for server: %s", err)
		return true
	}
	defer res.Body.Close()
	return err != nil
}

func TestCommandLineInterface(t *testing.T) {
	var osArgs = make([]string, len(os.Args))
	copy(osArgs, os.Args)
	cmd := cmdx.CommandExecuter{
		New: func() *cobra.Command {
			cp := *RootCmd
			return &cp
		},
	}

	// start server, and wait for the ports to be open
	cmd.ExecBackground(nil, os.Stdout, os.Stderr, "serve", "--disable-telemetry")
	var count = 0
	for ensureOpen(t, apiPort) && ensureOpen(t, proxyPort) {
		t.Logf("Port is not yet open, retrying attempt #%d..", count)
		count++
		if count > 50 {
			t.FailNow()
		}
		time.Sleep(100 * time.Millisecond)
	}

	for _, c := range []struct {
		args []string
	}{
		{args: []string{"rules", fmt.Sprintf("--endpoint=http://127.0.0.1:%d/", apiPort), "list"}},
		{args: []string{"rules", fmt.Sprintf("--endpoint=http://127.0.0.1:%d/", apiPort), "get", "test-rule-4"}},
		{args: []string{"health", fmt.Sprintf("--endpoint=http://127.0.0.1:%d/", apiPort), "alive"}},
		{args: []string{"health", fmt.Sprintf("--endpoint=http://127.0.0.1:%d/", apiPort), "ready"}},
		{args: []string{"credentials", "generate", "--alg", "RS256"}},
		{args: []string{"credentials", "generate", "--alg", "ES256"}},
		{args: []string{"credentials", "generate", "--alg", "HS256"}},
		{args: []string{"credentials", "generate", "--alg", "RS512"}},
	} {
		t.Run(fmt.Sprintf("command=%v", c.args), func(t *testing.T) {
			cmd.ExecNoErr(t, c.args...)
		})
	}
}
