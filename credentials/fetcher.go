package credentials

import (
	"context"
	"gopkg.in/square/go-jose.v2"
	"net/url"
)

type Fetcher interface {
	ResolveKey(ctx context.Context, locations []url.URL, kid string, use string) (*jose.JSONWebKey, error)
	ResolveSets(ctx context.Context, locations []url.URL) ([]jose.JSONWebKeySet, error)
}

type FetchRegistry interface {
	CredentialsFetcher() Fetcher
}
