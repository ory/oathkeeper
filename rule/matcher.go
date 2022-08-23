package rule

import (
	"context"
	"net/url"
)

type (
	Protocol int

	Matcher interface {
		Match(ctx context.Context, method string, u *url.URL, protocol Protocol) (*Rule, error)
	}
)

const (
	ProtocolHTTP Protocol = iota
	ProtocolGRPC
)
