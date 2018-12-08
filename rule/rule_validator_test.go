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

package rule

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
)

func TestValidateRule(t *testing.T) {
	v := ValidateRule(
		[]string{"an1", "an2"}, []string{"an1", "an2", "an3"},
		[]string{"az1", "az2"}, []string{"az1", "az2", "az3"},
		[]string{"ci1", "ci2"}, []string{"ci1", "ci2", "ci3"},
	)

	assertReason(t, v(&Rule{}), "from match.url field is not a valid url.")

	// assertReason(t, v(&Rule{
	//	 Match: RuleMatch{URL: "asdf"},
	// }), "from match.url field is not a valid url.")

	assertReason(t, v(&Rule{
		Match: RuleMatch{URL: "https://www.ory.sh", Methods: []string{"FOO"}},
	}), "from match.methods is not a valid HTTP method")

	// assertReason(t, v(&Rule{
	//	 Match: RuleMatch{URL: "https://www.ory.sh", Methods: []string{"POST"}},
	// }), "from upstream.url field is not a valid url.")

	assertReason(t, v(&Rule{
		Match:    RuleMatch{URL: "https://www.ory.sh", Methods: []string{"POST"}},
		Upstream: Upstream{URL: "foo"},
	}), "from upstream.url field is not a valid url.")

	assertReason(t, v(&Rule{
		Match:    RuleMatch{URL: "https://www.ory.sh", Methods: []string{"POST"}},
		Upstream: Upstream{URL: "https://www.ory.sh"},
	}), "At least one authenticator must be set.")

	assertReason(t, v(&Rule{
		Match:          RuleMatch{URL: "https://www.ory.sh", Methods: []string{"POST"}},
		Upstream:       Upstream{URL: "https://www.ory.sh"},
		Authenticators: []RuleHandler{{Handler: "foo"}},
	}), "is unknown, enabled authenticators are")

	assertReason(t, v(&Rule{
		Match:          RuleMatch{URL: "https://www.ory.sh", Methods: []string{"POST"}},
		Upstream:       Upstream{URL: "https://www.ory.sh"},
		Authenticators: []RuleHandler{{Handler: "an3"}},
	}), "is valid but has not enabled by the server's configuration, enabled authorizers are:")

	assertReason(t, v(&Rule{
		Match:          RuleMatch{URL: "https://www.ory.sh", Methods: []string{"POST"}},
		Upstream:       Upstream{URL: "https://www.ory.sh"},
		Authenticators: []RuleHandler{{Handler: "an1"}},
	}), "Value authorizer.handler can not be empty.")

	assertReason(t, v(&Rule{
		Match:          RuleMatch{URL: "https://www.ory.sh", Methods: []string{"POST"}},
		Upstream:       Upstream{URL: "https://www.ory.sh"},
		Authenticators: []RuleHandler{{Handler: "an1"}},
		Authorizer:     RuleHandler{Handler: "foo"},
	}), "s unknown, enabled authorizers are:")

	assertReason(t, v(&Rule{
		Match:          RuleMatch{URL: "https://www.ory.sh", Methods: []string{"POST"}},
		Upstream:       Upstream{URL: "https://www.ory.sh"},
		Authenticators: []RuleHandler{{Handler: "an1"}},
		Authorizer:     RuleHandler{Handler: "az2"},
	}), "Value credentials_issuer.handler can not be empty.")

	assertReason(t, v(&Rule{
		Match:             RuleMatch{URL: "https://www.ory.sh", Methods: []string{"POST"}},
		Upstream:          Upstream{URL: "https://www.ory.sh"},
		Authenticators:    []RuleHandler{{Handler: "an1"}},
		Authorizer:        RuleHandler{Handler: "az2"},
		CredentialsIssuer: RuleHandler{Handler: "foo"},
	}), "is unknown, enabled credentials issuers are:")

	assert.NoError(t, v(&Rule{
		Match:             RuleMatch{URL: "https://www.ory.sh", Methods: []string{"POST"}},
		Upstream:          Upstream{URL: "https://www.ory.sh"},
		Authenticators:    []RuleHandler{{Handler: "an1"}, {Handler: "an2"}},
		Authorizer:        RuleHandler{Handler: "az1"},
		CredentialsIssuer: RuleHandler{Handler: "ci1"},
	}))

	assert.NoError(t, v(&Rule{
		Match:             RuleMatch{URL: "https://www.ory.sh", Methods: []string{"POST"}},
		Upstream:          Upstream{URL: "https://www.ory.sh"},
		Authenticators:    []RuleHandler{{Handler: "an2"}},
		Authorizer:        RuleHandler{Handler: "az1"},
		CredentialsIssuer: RuleHandler{Handler: "ci2"},
	}))
}

func assertReason(t *testing.T, err error, sub string) {
	require.Error(t, err)
	reason := errors.Cause(err).(*herodot.DefaultError).ReasonField
	assert.True(t, strings.Contains(reason, sub), "%s", reason)
}
