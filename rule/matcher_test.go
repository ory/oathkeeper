// Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package rule

import (
	"fmt"
	"net/url"
	"strconv"
	"testing"

	"github.com/ory/ladon/compiler"
)

var methods = []string{"POST", "PUT", "GET", "DELETE", "PATCH", "OPTIONS", "HEAD"}

func generateDummyRules(amount int) []Rule {
	rules := make([]Rule, amount)
	scopes := []string{"foo", "bar", "baz", "faz"}
	expressions := []string{"/users/", "/users", "/blogs/", "/use<(r)>s/"}
	resources := []string{"users", "users:$1"}
	actions := []string{"get", "get:$1"}

	for i := 0; i < amount; i++ {
		exp, _ := compiler.CompileRegex(expressions[(i%(len(expressions)))]+"([0-"+strconv.Itoa(i)+"]+)", '<', '>')
		rules[i] = Rule{
			ID:                 strconv.Itoa(i),
			MatchesMethods:     methods[:i%(len(methods))],
			RequiredScopes:     scopes[:i%(len(scopes))],
			RequiredAction:     actions[i%(len(actions))],
			RequiredResource:   resources[i%(len(resources))],
			MatchesURLCompiled: exp,
		}
	}
	return rules
}

func generateUrls(amount int) []*url.URL {
	urls := make([]*url.URL, amount)
	for i := 0; i < amount; i++ {
		parsed, _ := url.Parse("/users/" + strconv.Itoa(i))
		urls[i] = parsed
	}
	return urls
}

func cachedRuleMatcherBenchmark(rules int) func(b *testing.B) {
	return func(b *testing.B) {
		b.StopTimer()
		matcher := &CachedMatcher{Rules: generateDummyRules(rules)}
		urls := generateUrls(10)
		methodLength := len(methods)
		urlsLength := len(urls)
		var matchedRules int

		b.StartTimer()
		for n := 0; n < b.N; n++ {
			_, err := matcher.MatchRule(
				methods[n%methodLength],
				urls[n%urlsLength],
			)
			if err != nil {
				matchedRules += 1
			}
		}
		b.StopTimer()
		b.Logf("Received %d rules", matchedRules)
	}
}

func BenchmarkCachedRuleMatcher(b *testing.B) {
	for _, tc := range []int{1, 3, 5, 7, 10, 100, 1000, 10000, 100000, 200000, 300000, 400000} {
		b.Run(fmt.Sprintf("rules=%d", tc), cachedRuleMatcherBenchmark(tc))
	}
}
