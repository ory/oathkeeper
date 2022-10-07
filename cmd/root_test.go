// Copyright © 2022 Ory Corp

package cmd

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/phayes/freeport"

	"github.com/stretchr/testify/assert"
)

var apiPort, proxyPort int

func freePort() (int, int) {
	var err error
	r := make([]int, 2)

	if r[0], err = freeport.GetFreePort(); err != nil {
		panic(err.Error())
	}

	tries := 0
	for {
		r[1], err = freeport.GetFreePort()
		if r[0] != r[1] {
			break
		}
		tries++
		if tries > 10 {
			panic("Unable to find free port")
		}
	}
	return r[0], r[1]
}

func init() {
	apiPort, proxyPort = freePort()

	os.Setenv("SERVE_API_PORT", fmt.Sprintf("%d", apiPort))
	os.Setenv("SERVE_PROXY_PORT", fmt.Sprintf("%d", proxyPort))
	os.Setenv("AUTHENTICATORS_NOOP_ENABLED", "1")
	os.Setenv("AUTHENTICATORS_ANONYMOUS_ENABLED", "true")
	os.Setenv("AUTHORIZERS_ALLOW_ENABLED", "true")
	os.Setenv("MUTATORS_NOOP_ENABLED", "true")
	os.Setenv("ACCESS_RULES_REPOSITORIES", "inline://W3siaWQiOiJ0ZXN0LXJ1bGUtNCIsInVwc3RyZWFtIjp7InByZXNlcnZlX2hvc3QiOnRydWUsInN0cmlwX3BhdGgiOiIvYXBpIiwidXJsIjoibXliYWNrZW5kLmNvbS9hcGkifSwibWF0Y2giOnsidXJsIjoibXlwcm94eS5jb20vYXBpIiwibWV0aG9kcyI6WyJHRVQiLCJQT1NUIl19LCJhdXRoZW50aWNhdG9ycyI6W3siaGFuZGxlciI6Im5vb3AifSx7ImhhbmRsZXIiOiJhbm9ueW1vdXMifV0sImF1dGhvcml6ZXIiOnsiaGFuZGxlciI6ImFsbG93In0sIm11dGF0b3JzIjpbeyJoYW5kbGVyIjoibm9vcCJ9XX1d")
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

	for _, c := range []struct {
		args      []string
		wait      func() bool
		expectErr bool
	}{
		{
			args: []string{"serve", "--disable-telemetry"},
			wait: func() bool {
				return ensureOpen(t, apiPort) && ensureOpen(t, proxyPort)
			},
		},
		{args: []string{"rules", fmt.Sprintf("--endpoint=http://127.0.0.1:%d/", apiPort), "list"}},
		{args: []string{"rules", fmt.Sprintf("--endpoint=http://127.0.0.1:%d/", apiPort), "get", "test-rule-4"}},
		{args: []string{"health", fmt.Sprintf("--endpoint=http://127.0.0.1:%d/", apiPort), "alive"}},
		{args: []string{"health", fmt.Sprintf("--endpoint=http://127.0.0.1:%d/", apiPort), "ready"}},
		{args: []string{"credentials", "generate", "--alg", "RS256"}},
		{args: []string{"credentials", "generate", "--alg", "ES256"}},
		{args: []string{"credentials", "generate", "--alg", "HS256"}},
		{args: []string{"credentials", "generate", "--alg", "RS512"}},
	} {
		RootCmd.SetArgs(c.args)

		t.Run(fmt.Sprintf("command=%v", c.args), func(t *testing.T) {
			if c.wait != nil {
				go func() {
					assert.Nil(t, RootCmd.Execute())
				}()
			}

			if c.wait != nil {
				var count = 0
				for c.wait() {
					t.Logf("Port is not yet open, retrying attempt #%d..", count)
					count++
					if count > 5 {
						t.FailNow()
					}
					time.Sleep(time.Second)
				}
			} else {
				err := RootCmd.Execute()
				if c.expectErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}
