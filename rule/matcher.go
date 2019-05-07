package rule

import (
	"context"
	"net/url"
)

type Matcher interface {
	Match(ctx context.Context, method string, u *url.URL) (*Rule, error)
}
