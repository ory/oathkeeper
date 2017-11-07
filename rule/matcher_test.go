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
			ID:                  strconv.Itoa(i),
			MatchesMethods:      methods[:i%(len(methods))],
			RequiredScopes:      scopes[:i%(len(scopes))],
			RequiredAction:      actions[i%(len(actions))],
			RequiredResource:    resources[i%(len(resources))],
			MatchesPathCompiled: exp,
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
