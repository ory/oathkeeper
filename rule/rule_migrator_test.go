// Copyright 2021 Ory GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/x"
)

func TestRuleMigration(t *testing.T) {
	for k, tc := range []struct {
		d         string
		in        string
		out       string
		expectErr bool
		version   string
	}{
		{
			d:       "should work with v0.19.0-beta.1",
			in:      `{}`,
			out:     `{"id":"","version":"v0.19.0-beta.1","description":"","match":null,"errors":null,"authenticators":null,"authorizer":{"handler":"","config":null},"mutators":null,"upstream":{"preserve_host":false,"strip_path":"","url":""}}`,
			version: "v0.19.0-beta.1",
		},
		{
			d:       "should work with v0.19.0-beta.1+oryOS.12",
			in:      `{}`,
			out:     `{"id":"","version":"v0.19.0-beta.1","description":"","match":null,"errors":null,"authenticators":null,"authorizer":{"handler":"","config":null},"mutators":null,"upstream":{"preserve_host":false,"strip_path":"","url":""}}`,
			version: "v0.19.0-beta.1+oryOS.12",
		},
		{
			d:       "should work with v0.19.0-beta.1",
			in:      `{"version":"v0.19.0-beta.1"}`,
			out:     `{"id":"","version":"v0.19.0-beta.1","description":"","match":null,"errors":null,"authenticators":null,"authorizer":{"handler":"","config":null},"mutators":null,"upstream":{"preserve_host":false,"strip_path":"","url":""}}`,
			version: "v0.19.0-beta.1",
		},
		{
			d:       "should work with 0.19.0-beta.1",
			in:      `{"version":"0.19.0-beta.1"}`,
			out:     `{"id":"","version":"v0.19.0-beta.1","description":"","match":null,"errors":null,"authenticators":null,"authorizer":{"handler":"","config":null},"mutators":null,"upstream":{"preserve_host":false,"strip_path":"","url":""}}`,
			version: "v0.19.0-beta.1+oryOS.12",
		},
		{
			d: "should migrate to 0.33.0",
			in: `{
  "version": "v0.30.0-beta.1",
  "mutators": [
	{},	
    {
      "handler": "hydrator",
      "config": {
        "retry": {
          "delay_in_milliseconds": 500,
          "number_of_retries": 5
        }
      }
    }
  ]
}`,
			out: `{
  "id": "",
  "version": "v0.33.0-beta.1",
  "description":"","match":null,"authenticators":null,"authorizer":{"handler":"","config":null},"errors":null,
  "mutators": [
	{"handler":"","config":null},
    {
      "handler": "hydrator",
      "config": {
        "retry": {
          "max_delay": "500ms",
          "give_up_after": "2500ms"
        }
      }
    }
  ],
  "upstream":{"preserve_host":false,"strip_path":"","url":""}
}`,
			version: "v0.33.0-beta.1+oryOS.12",
		},
		{
			d: "should migrate to 0.37.0",
			in: `{
				"version": "v0.33.0",
				"authorizer":	
				  {
					"handler": "keto_engine_acp_ory",
					"config": {
						"required_action": "my:action:$1",
						"required_resource": "my:resource:$2:foo:$1",
						"flavor": "exact"
					}
				  }
			  }`,
			out: `{
				"id": "",
				"version": "v0.37.0",
				"description":"","match":null,"authenticators":null,"errors":null,
				"authorizer":	
				  {
					"handler": "keto_engine_acp_ory",
					"config": {
						"required_action": "my:action:{{ printIndex .MatchContext.RegexpCaptureGroups (sub 1 1 | int)}}",
						"required_resource": "my:resource:{{ printIndex .MatchContext.RegexpCaptureGroups (sub 2 1 | int)}}:foo:{{ printIndex .MatchContext.RegexpCaptureGroups (sub 1 1 | int)}}",
						"flavor": "exact"
					}
				  },
				"mutators": null,
				"upstream":{"preserve_host":false,"strip_path":"","url":""}
			  }`,
			version: "v0.37.0+oryOS.18",
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.d), func(t *testing.T) {
			var r Rule

			x.Version = tc.version
			err := json.NewDecoder(bytes.NewBufferString(tc.in)).Decode(&r)
			x.Version = x.UnknownVersion

			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err, "%+v", err)

			var out bytes.Buffer
			require.NoError(t, json.NewEncoder(&out).Encode(&r))
			assert.JSONEq(t, tc.out, out.String(), "%s", out.String())
		})
	}
}
