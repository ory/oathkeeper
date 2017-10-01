package rule

import (
	"net/url"
)

type Matcher interface {
	MatchRules(method string, u *url.URL) ([]Rule, error)
}
