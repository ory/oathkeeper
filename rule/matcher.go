package rule

import (
	"context"
	"net/http"
	"net/url"
)

type Matcher interface {
	Match(ctx context.Context, method string, u *url.URL, headers http.Header) (*Rule, error)
}
