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

package rule_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ory/viper"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/internal"
	. "github.com/ory/oathkeeper/rule"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
)

func TestValidateRule(t *testing.T) {
	var prep = func(an, az, m bool) func() {
		return func() {
			viper.Set(configuration.ViperKeyAuthenticatorNoopIsEnabled, an)
			viper.Set(configuration.ViperKeyAuthorizerAllowIsEnabled, az)
			viper.Set(configuration.ViperKeyMutatorNoopIsEnabled, m)
		}
	}

	for k, tc := range []struct {
		setup     func()
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
				tc.setup()
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

func assertReason(t *testing.T, err error, sub string) {
	require.Error(t, err)
	reason := errors.Cause(err).(*herodot.DefaultError).ReasonField
	assert.True(t, strings.Contains(reason, sub), "%s", reason)
}
