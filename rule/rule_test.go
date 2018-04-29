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
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustParse(t *testing.T, u string) *url.URL {
	p, err := url.Parse(u)
	require.NoError(t, err)
	return p
}

func TestRule(t *testing.T) {
	r := &Rule{
		Match: RuleMatch{
			Methods: []string{"DELETE"},
			URL:     "https://localhost/users/<[0-9]+>",
		},
	}

	assert.NoError(t, r.IsMatching("DELETE", mustParse(t, "https://localhost/users/1234")))
	assert.Error(t, r.IsMatching("DELETE", mustParse(t, "https://localhost/users/abcd")))
}
