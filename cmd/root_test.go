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
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCommandLineInterface(t *testing.T) {
	var osArgs = make([]string, len(os.Args))
	os.Setenv("PORT", "4456")
	os.Setenv("DATABASE_URL", "memory")
	os.Setenv("CREDENTIALS_ISSUER_ID_TOKEN_HS256_SECRET", "dYmTueb6zg8TphfZbOUpOewd0gt7u0SH")
	copy(osArgs, os.Args)

	for _, c := range []struct {
		args      []string
		wait      func() bool
		expectErr bool
	}{
		{
			args: []string{"serve", "--disable-telemetry", "api"},
			wait: func() bool {
				res, err := http.Get("http://localhost:4456")
				if err != nil {
					t.Logf("Network error while polling for server: %s", err)
				} else {
					defer res.Body.Close()
				}
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
