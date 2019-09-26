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
			out:     `{"id":"","version":"v0.19.0-beta.1","description":"","match":{"methods":null,"url":""},"authenticators":null,"authorizer":{"handler":"","config":null},"mutators":null,"upstream":{"preserve_host":false,"strip_path":"","url":""}}`,
			version: "v0.19.0-beta.1",
		},
		{
			d:       "should work with v0.19.0-beta.1+oryOS.12",
			in:      `{}`,
			out:     `{"id":"","version":"v0.19.0-beta.1","description":"","match":{"methods":null,"url":""},"authenticators":null,"authorizer":{"handler":"","config":null},"mutators":null,"upstream":{"preserve_host":false,"strip_path":"","url":""}}`,
			version: "v0.19.0-beta.1+oryOS.12",
		},
		{
			d:       "should work with v0.19.0-beta.1",
			in:      `{"version":"v0.19.0-beta.1"}`,
			out:     `{"id":"","version":"v0.19.0-beta.1","description":"","match":{"methods":null,"url":""},"authenticators":null,"authorizer":{"handler":"","config":null},"mutators":null,"upstream":{"preserve_host":false,"strip_path":"","url":""}}`,
			version: "v0.19.0-beta.1",
		},
		{
			d:       "should work with 0.19.0-beta.1",
			in:      `{"version":"0.19.0-beta.1"}`,
			out:     `{"id":"","version":"v0.19.0-beta.1","description":"","match":{"methods":null,"url":""},"authenticators":null,"authorizer":{"handler":"","config":null},"mutators":null,"upstream":{"preserve_host":false,"strip_path":"","url":""}}`,
			version: "v0.19.0-beta.1+oryOS.12",
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
