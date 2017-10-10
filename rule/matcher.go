package rule

import (
	"net/url"
)

type Matcher interface {
	MatchRule(method string, u *url.URL) (*Rule, error)
}
