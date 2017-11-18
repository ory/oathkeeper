package cmd

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommandLineInterface(t *testing.T) {
	var osArgs = make([]string, len(os.Args))
	os.Setenv("DATABASE_URL", "memory")
	os.Setenv("HYDRA_URL", "http://does-not-exist.com/")
	os.Setenv("HYDRA_CLIENT_ID", "does-not-exist")
	os.Setenv("HYDRA_CLIENT_SECRET", "does-not-exist")
	copy(osArgs, os.Args)

	for _, c := range []struct {
		args      []string
		wait      func() bool
		expectErr bool
	}{
		{
			args: []string{"serve", "all"},
			wait: func() bool {
				res, err := http.Get("http://localhost:4455")
				if err != nil {
					t.Logf("Network error while polling for server: %s", err)
				}
				defer res.Body.Close()
				return err != nil
			},
		},
		{args: []string{"rules", "--endpoint=http://127.0.0.1:4456/", "import", "../stub/rules.json"}},
		{args: []string{"rules", "--endpoint=http://127.0.0.1:4456/", "list"}},
		{args: []string{"rules", "--endpoint=http://127.0.0.1:4456/", "get", "test-rule-1"}},
		{args: []string{"rules", "--endpoint=http://127.0.0.1:4456/", "get", "test-rule-2"}},
		{args: []string{"rules", "--endpoint=http://127.0.0.1:4456/", "delete", "test-rule-1"}},
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
					t.Logf("Port is not yet open, retrying attempt #%d...", count)
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
