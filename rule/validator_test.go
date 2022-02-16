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
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ory/oathkeeper/driver"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/driver/configuration"
	. "github.com/ory/oathkeeper/rule"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
)

func TestValidateRule(t *testing.T) {
	for k, tc := range []struct {
		configOpts func() []configx.OptionModifier
		r          *Rule
		expectErr  string
	}{
		{
			configOpts: func() []configx.OptionModifier { return nil },
			r:          &Rule{},
			expectErr:  `Value "match" is empty but must be set.`,
		},
		{
			configOpts: func() []configx.OptionModifier { return nil },
			r:          &Rule{Match: &Match{}},
			expectErr:  `Value "" of "match.url" field is not a valid url.`,
		},
		{
			configOpts: func() []configx.OptionModifier { return nil },
			r: &Rule{
				Match:    &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream: Upstream{URL: "https://www.ory.sh"},
			},
			expectErr: `Value of "authenticators" must be set and can not be an empty array.`,
		},
		{
			configOpts: func() []configx.OptionModifier {
				return []configx.OptionModifier{
					configx.WithValue(configuration.ViperKeyAuthenticatorNoopIsEnabled, true),
					configx.WithValue(configuration.ViperKeyAuthorizerAllowIsEnabled, false),
					configx.WithValue(configuration.ViperKeyMutatorNoopIsEnabled, false),
				}
			},
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "foo"}},
			},
			expectErr: `Value "foo" of "authenticators[0]" is not in list of supported authenticators: `,
		},
		{
			configOpts: func() []configx.OptionModifier {
				return []configx.OptionModifier{
					configx.WithValue(configuration.ViperKeyAuthenticatorNoopIsEnabled, false),
					configx.WithValue(configuration.ViperKeyAuthorizerAllowIsEnabled, false),
					configx.WithValue(configuration.ViperKeyMutatorNoopIsEnabled, false),
				}
			},
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
			},
			expectErr: `Authenticator "noop" is disabled per configuration.`,
		},
		{
			configOpts: func() []configx.OptionModifier {
				return []configx.OptionModifier{
					configx.WithValue(configuration.ViperKeyAuthenticatorNoopIsEnabled, true),
					configx.WithValue(configuration.ViperKeyAuthorizerAllowIsEnabled, false),
					configx.WithValue(configuration.ViperKeyMutatorNoopIsEnabled, false),
				}
			},
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
			},
			expectErr: `Value of "authorizer.handler" can not be empty.`,
		},
		{
			configOpts: func() []configx.OptionModifier {
				return []configx.OptionModifier{
					configx.WithValue(configuration.ViperKeyAuthenticatorNoopIsEnabled, true),
					configx.WithValue(configuration.ViperKeyAuthorizerAllowIsEnabled, true),
					configx.WithValue(configuration.ViperKeyMutatorNoopIsEnabled, false),
				}
			},
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
				Authorizer:     Handler{Handler: "foo"},
			},
			expectErr: `Value "foo" of "authorizer.handler" is not in list of supported authorizers: `,
		},
		{
			configOpts: func() []configx.OptionModifier {
				return []configx.OptionModifier{
					configx.WithValue(configuration.ViperKeyAuthenticatorNoopIsEnabled, true),
					configx.WithValue(configuration.ViperKeyAuthorizerAllowIsEnabled, true),
					configx.WithValue(configuration.ViperKeyMutatorNoopIsEnabled, false),
				}
			},
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
				Authorizer:     Handler{Handler: "allow"},
			},
			expectErr: `Value of "mutators" must be set and can not be an empty array.`,
		},
		{
			configOpts: func() []configx.OptionModifier {
				return []configx.OptionModifier{
					configx.WithValue(configuration.ViperKeyAuthenticatorNoopIsEnabled, true),
					configx.WithValue(configuration.ViperKeyAuthorizerAllowIsEnabled, true),
					configx.WithValue(configuration.ViperKeyMutatorNoopIsEnabled, true),
				}
			},
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
			configOpts: func() []configx.OptionModifier {
				return []configx.OptionModifier{
					configx.WithValue(configuration.ViperKeyAuthenticatorNoopIsEnabled, true),
					configx.WithValue(configuration.ViperKeyAuthorizerAllowIsEnabled, true),
					configx.WithValue(configuration.ViperKeyMutatorNoopIsEnabled, true),
				}
			},
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"POST"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
				Authorizer:     Handler{Handler: "allow"},
				Mutators:       []Handler{{Handler: "noop"}},
			},
		},
		{
			configOpts: func() []configx.OptionModifier {
				return []configx.OptionModifier{
					configx.WithValue(configuration.ViperKeyAuthenticatorNoopIsEnabled, true),
					configx.WithValue(configuration.ViperKeyAuthorizerAllowIsEnabled, true),
					configx.WithValue(configuration.ViperKeyMutatorNoopIsEnabled, true),
				}
			},
			r: &Rule{
				Match:          &Match{URL: "https://www.ory.sh", Methods: []string{"MKCOL"}},
				Upstream:       Upstream{URL: "https://www.ory.sh"},
				Authenticators: []Handler{{Handler: "noop"}},
				Authorizer:     Handler{Handler: "allow"},
				Mutators:       []Handler{{Handler: "noop"}},
			},
		},
		{
			configOpts: func() []configx.OptionModifier {
				return []configx.OptionModifier{
					configx.WithValue(configuration.ViperKeyAuthenticatorNoopIsEnabled, true),
					configx.WithValue(configuration.ViperKeyAuthorizerAllowIsEnabled, true),
					configx.WithValue(configuration.ViperKeyMutatorNoopIsEnabled, false),
				}
			},
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
			conf, err := configuration.NewViperProvider(context.Background(), logrusx.New("", ""),
				append(tc.configOpts(),
					configx.WithValue("log.level", "debug"),
					configx.WithValue(configuration.ViperKeyErrorsJSONIsEnabled, true))...)
			require.NoError(t, err)

			r := driver.NewRegistryMemory().WithConfig(conf)
			v := NewValidatorDefault(r)

			err = v.Validate(tc.r)
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
