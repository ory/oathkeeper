// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	. "github.com/ory/oathkeeper/rule"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
)

func TestValidateRule(t *testing.T) {
	var prep = func(an, az, m bool) func(*testing.T, configuration.Provider) {
		return func(t *testing.T, config configuration.Provider) {
			config.SetForTest(t, configuration.AuthenticatorNoopIsEnabled, an)
			config.SetForTest(t, configuration.AuthorizerAllowIsEnabled, az)
			config.SetForTest(t, configuration.MutatorNoopIsEnabled, m)
		}
	}

	for k, tc := range []struct {
		setup     func(*testing.T, configuration.Provider)
		r         *Rule
		expectErr string
	}{
		{
			r:         &Rule{},
			expectErr: `Value "match" is empty but must be set.`,
		},
		{
			r:         &Rule{Match: &Match{}},
			expectErr: `Value "" of "match.url" field must not be empty.`,
		},
		{
			r: &Rule{
				Match:    &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream: Upstream{URL: "https://www.ory.sh"},
			},
			expectErr: `Value of "authenticators" must be set and can not be an empty array.`,
		},
		{
			setup: prep(true, false, false),
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "foo"}},
			},
			expectErr: `Value "foo" of "authenticators[0]" is not in list of supported authenticators: `,
		},
		{
			setup: prep(false, false, false),
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
			},
			expectErr: `Authenticator "noop" is disabled per configuration.`,
		},
		{
			setup: prep(true, false, false),
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
			},
			expectErr: `Value of "authorizer.handler" can not be empty.`,
		},
		{
			setup: prep(true, true, false),
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
				Authorizer:     Handler{Handler: "foo"},
			},
			expectErr: `Value "foo" of "authorizer.handler" is not in list of supported authorizers: `,
		},
		{
			setup: prep(true, true, false),
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
				Authorizer:     Handler{Handler: "allow"},
			},
			expectErr: `Value of "mutators" must be set and can not be an empty array.`,
		},
		{
			setup: prep(true, true, true),
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
				Authorizer:     Handler{Handler: "allow"},
				Mutators:       []Handler{{Handler: "foo"}},
			},
			expectErr: `Value "foo" of "mutators[0]" is not in list of supported mutators: `,
		},
		{
			setup: prep(true, true, true),
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
				Authorizer:     Handler{Handler: "allow"},
				Mutators:       []Handler{{Handler: "noop"}},
			},
		},
		{
			setup: prep(true, true, true),
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"MKCOL"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
				Authorizer:     Handler{Handler: "allow"},
				Mutators:       []Handler{{Handler: "noop"}},
			},
		},
		{
			setup: prep(true, true, true),
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"MKCOL"}},
				Upstream:       Upstream{URL: "http://tasks.foo-bar_baz"},
				Authenticators: []Handler{{Handler: "noop"}},
				Authorizer:     Handler{Handler: "allow"},
				Mutators:       []Handler{{Handler: "noop"}},
			},
		},
		{
			setup: prep(true, true, false),
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
				Authorizer:     Handler{Handler: "allow"},
				Mutators:       []Handler{{Handler: "noop"}},
			},
			expectErr: `Mutator "noop" is disabled per configuration.`,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			conf := internal.NewConfigurationWithDefaults()
			if tc.setup != nil {
				tc.setup(t, conf)
			}

			r := internal.NewRegistry(conf)
			v := NewValidatorDefault(r)

			err := v.Validate(tc.r)
			if tc.expectErr == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			reason := errors.Cause(err).(*herodot.DefaultError).ReasonField
			assert.True(t, strings.Contains(reason, tc.expectErr), "%s != %s", reason, tc.expectErr)
		})
	}
}
