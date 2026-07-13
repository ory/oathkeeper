// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package credentials

import (
	"context"
	"encoding/json"
	stderrs "errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/pkg/errors"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
	"github.com/ory/x/logrusx"
	"github.com/ory/x/otelx"
	"github.com/ory/x/urlx"
)

type reasoner interface {
	error
	Reason() string
}

var _ Fetcher = new(FetcherDefault)

type FetcherDefault struct {
	sync.RWMutex

	ttl, cancelAfter time.Duration
	client           *http.Client
	keys             map[string]jose.JSONWebKeySet
	fetchedAt        map[string]time.Time
	l                *logrusx.Logger
	mux              *blob.URLMux
}

type dependencies interface {
	logrusx.Provider
	otelx.Provider
}

type FetcherOption func(f *FetcherDefault)

func WithURLMux(mux *blob.URLMux) FetcherOption {
	return func(f *FetcherDefault) { f.mux = mux }
}

// NewFetcherDefault returns a new JWKS Fetcher with:
//
//   - cancelAfter: If reached, the fetcher will stop waiting for responses and return an error.
//   - waitForResponse: While the fetcher might stop waiting for responses, we will give the server more time to respond
//     and add the keys to the registry unless waitForResponse is reached in which case we'll terminate the request.
func NewFetcherDefault(d dependencies, cancelAfter, ttl time.Duration, opts ...FetcherOption) *FetcherDefault {
	f := &FetcherDefault{
		cancelAfter: cancelAfter,
		l:           d.Logger(),
		ttl:         ttl,
		keys:        make(map[string]jose.JSONWebKeySet),
		fetchedAt:   make(map[string]time.Time),
		client: httpx.NewResilientClient(
			httpx.ResilientClientWithConnectionTimeout(15 * time.Second),
		).StandardClient(),
		mux: blob.DefaultURLMux(),
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

func (s *FetcherDefault) ResolveSets(ctx context.Context, locations []url.URL) ([]jose.JSONWebKeySet, error) {
	if set := s.set(locations, false); set != nil {
		return set, nil
	}

	fetchError := s.fetchParallel(ctx, locations)

	if set := s.set(locations, errors.Is(fetchError, context.DeadlineExceeded)); set != nil {
		return set, nil
	}

	return nil, errors.WithStack(herodot.ErrInternalServerError().
		WithReasonf(`None of the provided URLs returned a valid JSON Web Key Set.`),
	)
}

func (s *FetcherDefault) fetchParallel(ctx context.Context, locations []url.URL) error {
	done := make(chan struct{})

	go s.resolveAll(ctx, done, locations)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(s.cancelAfter):
		s.l.Errorf("Ignoring JSON Web Keys from at least one URI because the request timed out waiting for a response.")
		return context.DeadlineExceeded
	case <-done:
		// We're done!
		return nil
	}
}

func (s *FetcherDefault) ResolveKey(ctx context.Context, locations []url.URL, kid string, use string) (*jose.JSONWebKey, error) {
	if key := s.key(kid, locations, use, false); key != nil {
		return key, nil
	}

	fetchError := s.fetchParallel(ctx, locations)

	if key := s.key(kid, locations, use, errors.Is(fetchError, context.DeadlineExceeded)); key != nil {
		return key, nil
	}

	return nil, errors.WithStack(herodot.ErrInternalServerError().
		WithDetail("jwks_urls", fmt.Sprintf("%v", locations)).
		WithReasonf(`JSON Web Key ID "%s" with use "%s" could not be found in any of the provided URIs.`, kid, use).
		WithDebug("Check that the provided JSON Web Key URIs contain a key that can verify the signature of the provided JSON Web Key ID."),
	)
}

func (s *FetcherDefault) key(kid string, locations []url.URL, use string, staleKeyAcceptable bool) *jose.JSONWebKey {
	for _, l := range locations {
		s.RLock()
		keys, ok1 := s.keys[l.String()]
		fetchedAt, ok2 := s.fetchedAt[l.String()]
		s.RUnlock()

		if !ok1 || !ok2 || s.isKeyExpired(staleKeyAcceptable, fetchedAt) {
			continue
		}

		for _, k := range keys.Key(kid) {
			if k.Use == use {
				return &k
			}
		}
	}

	return nil
}

func (s *FetcherDefault) set(locations []url.URL, staleKeyAcceptable bool) []jose.JSONWebKeySet {
	var result []jose.JSONWebKeySet
	for _, l := range locations {
		s.RLock()
		keys, ok1 := s.keys[l.String()]
		fetchedAt, ok2 := s.fetchedAt[l.String()]
		s.RUnlock()

		if !ok1 || !ok2 || s.isKeyExpired(staleKeyAcceptable, fetchedAt) {
			continue
		}

		result = append(result, keys)
	}

	return result
}

func (s *FetcherDefault) isKeyExpired(expiredKeyAcceptable bool, fetchedAt time.Time) bool {
	return !expiredKeyAcceptable && time.Since(fetchedAt) > s.ttl
}

func (s *FetcherDefault) resolveAll(ctx context.Context, done chan<- struct{}, locations []url.URL) {
	// we don't want to cancel so the cache gets populated
	ctx = context.WithoutCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(len(locations))

	for _, l := range locations {
		wg.Go(func() {
			defer wg.Done()
			err := s.resolve(ctx, l)
			if err != nil {
				var reason string
				if r, ok := stderrs.AsType[reasoner](err); ok {
					reason = r.Reason()
				}
				s.l.WithError(err).
					WithField("stack", fmt.Sprintf("%+v", err)).
					WithField("reason", reason).
					Errorf("Unable to fetch JSON Web Key Set from remote")
			}
		})
	}

	wg.Wait()
	close(done)
}

func (s *FetcherDefault) resolve(ctx context.Context, location url.URL) error {
	var (
		reader io.ReadCloser
		err    error
	)

	switch location.Scheme {
	case "azblob", "gs", "s3":
		bucket, err := s.mux.OpenBucket(ctx, location.Scheme+"://"+location.Host)
		if err != nil {
			return errors.WithStack(herodot.ErrInternalServerError().
				WithReasonf(
					`Unable to fetch JSON Web Keys from location "%s" because "%s".`,
					location.String(),
					err,
				),
			)
		}
		defer func() { _ = bucket.Close() }()

		reader, err = bucket.NewReader(ctx, location.Path[1:], nil)
		if err != nil {
			return errors.WithStack(herodot.ErrInternalServerError().
				WithReasonf(
					`Unable to fetch JSON Web Keys from location "%s" because "%s".`,
					location.String(),
					err,
				),
			)
		}
		defer func() { _ = reader.Close() }()

	case "", "file":
		reader, err = os.Open(urlx.GetURLFilePath(&location))
		if err != nil {
			return errors.WithStack(herodot.ErrInternalServerError().
				WithReasonf(
					`Unable to fetch JSON Web Keys from location "%s" because "%s".`,
					location.String(),
					err,
				),
			)
		}
		defer func() { _ = reader.Close() }()

	case "http", "https":
		req, err := http.NewRequestWithContext(ctx, "GET", location.String(), nil)
		if err != nil {
			return errors.WithStack(herodot.ErrInternalServerError().
				WithReasonf(
					`Unable to fetch JSON Web Keys from location "%s" because "%s".`,
					location.String(),
					err,
				),
			)
		}
		res, err := s.client.Do(req)
		if err != nil {
			return errors.WithStack(herodot.ErrInternalServerError().
				WithReasonf(
					`Unable to fetch JSON Web Keys from location "%s" because "%s".`,
					location.String(),
					err,
				),
			)
		}
		reader = res.Body
		defer func() { _ = reader.Close() }()

		if res.StatusCode < 200 || res.StatusCode >= 400 {
			return errors.WithStack(herodot.ErrInternalServerError().
				WithReasonf(
					`Expected successful status code from location "%s", but received code "%d".`,
					location.String(),
					res.StatusCode,
				),
			)
		}

	default:
		return errors.WithStack(herodot.ErrInternalServerError().
			WithReasonf(
				`Unable to fetch JSON Web Keys from location "%s" because URL scheme "%s" is not supported.`,
				location.String(),
				location.Scheme,
			),
		)
	}

	var set jose.JSONWebKeySet
	if err := json.NewDecoder(reader).Decode(&set); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError().
			WithReasonf(
				`Unable to decode JSON Web Keys from location "%s" because "%s".`,
				location.String(),
				err,
			),
		)
	}

	s.Lock()
	s.keys[location.String()] = set
	s.fetchedAt[location.String()] = time.Now().UTC()
	s.Unlock()

	return nil
}
