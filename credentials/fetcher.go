// Copyright © 2022 Ory Corp

package credentials

import (
	"context"
	"net/url"

	"gopkg.in/square/go-jose.v2"
)

type Fetcher interface {
	ResolveKey(ctx context.Context, locations []url.URL, kid string, use string) (*jose.JSONWebKey, error)
	ResolveSets(ctx context.Context, locations []url.URL) ([]jose.JSONWebKeySet, error)
}

type FetcherRegistry interface {
	CredentialsFetcher() Fetcher
}
