/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author		Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @Copyright 	2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license 	Apache-2.0
 *
 */
package credentials

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/ory/oathkeeper/internal/cloudstorage"

	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/urlx"

	"github.com/ory/herodot"
	"github.com/ory/x/httpx"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
)

type reasoner interface {
	Reason() string
}

var _ Fetcher = new(FetcherDefault)

type FetcherDefault struct {
	sync.RWMutex

	ttl         time.Duration
	cancelAfter time.Duration
	client      *http.Client
	keys        map[string]jose.JSONWebKeySet
	fetchedAt   map[string]time.Time
	l           *logrusx.Logger
	mux         *blob.URLMux
}

// NewFetcherDefault returns a new JWKS Fetcher with:
//
// - cancelAfter: If reached, the fetcher will stop waiting for responses and return an error.
// - waitForResponse: While the fetcher might stop waiting for responses, we will give the server more time to respond
//		and add the keys to the registry unless waitForResponse is reached in which case we'll terminate the request.
func NewFetcherDefault(l *logrusx.Logger, cancelAfter time.Duration, ttl time.Duration) *FetcherDefault {
	return &FetcherDefault{
		cancelAfter: cancelAfter,
		l:           l,
		ttl:         ttl,
		keys:        make(map[string]jose.JSONWebKeySet),
		fetchedAt:   make(map[string]time.Time),
		client:      httpx.NewResilientClientLatencyToleranceHigh(nil),
		mux:         cloudstorage.NewURLMux(),
	}
}

func (s *FetcherDefault) ResolveSets(ctx context.Context, locations []url.URL) ([]jose.JSONWebKeySet, error) {
	if set := s.set(locations, false); set != nil {
		return set, nil
	}

	fetchError := s.fetchParallel(ctx, locations)

	if set := s.set(locations, errors.Is(fetchError, context.DeadlineExceeded)); set != nil {
		return set, nil
	}

	return nil, errors.WithStack(herodot.
		ErrInternalServerError.
		WithReasonf(`None of the provided URLs returned a valid JSON Web Key Set.`),
	)
}

func (s *FetcherDefault) fetchParallel(ctx context.Context, locations []url.URL) error {
	ctx, cancel := context.WithTimeout(ctx, s.cancelAfter)
	defer cancel()
	errs := make(chan error)
	done := make(chan struct{})

	go func() {
		for err := range errs {
			var reason string
			if r, ok := errors.Cause(err).(reasoner); ok {
				reason = r.Reason()
			}
			s.l.WithError(err).
				WithField("stack", fmt.Sprintf("%+v", err)).
				WithField("reason", reason).
				Errorf("Unable to fetch JSON Web Key Set from remote")
		}
	}()

	go s.resolveAll(done, errs, locations)

	select {
	case <-ctx.Done():
		s.l.WithError(ctx.Err()).Errorf("Ignoring JSON Web Keys from at least one URI because the request timed out waiting for a response.")
		return ctx.Err()
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

	return nil, errors.WithStack(herodot.
		ErrInternalServerError.
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
	return expiredKeyAcceptable == false &&
		fetchedAt.Add(s.ttl).Before(time.Now().UTC())
}

func (s *FetcherDefault) resolveAll(done chan struct{}, errs chan error, locations []url.URL) {
	var wg sync.WaitGroup

	for _, l := range locations {
		wg.Add(1)
		go s.resolve(&wg, errs, l)
	}

	wg.Wait()
	close(done)
	close(errs)
}

func (s *FetcherDefault) resolve(wg *sync.WaitGroup, errs chan error, location url.URL) {
	defer wg.Done()
	var reader io.Reader

	switch location.Scheme {
	case "azblob":
		fallthrough
	case "gs":
		fallthrough
	case "s3":
		ctx := context.Background()
		bucket, err := s.mux.OpenBucket(ctx, location.Scheme+"://"+location.Host)
		if err != nil {
			errs <- errors.WithStack(herodot.
				ErrInternalServerError.
				WithReasonf(
					`Unable to fetch JSON Web Keys from location "%s" because "%s".`,
					location.String(),
					err,
				),
			)
			return
		}
		defer bucket.Close()

		r, err := bucket.NewReader(ctx, location.Path[1:], nil)
		if err != nil {
			errs <- errors.WithStack(herodot.
				ErrInternalServerError.
				WithReasonf(
					`Unable to fetch JSON Web Keys from location "%s" because "%s".`,
					location.String(),
					err,
				),
			)
			return
		}
		defer r.Close()

		reader = r
	case "":
		fallthrough
	case "file":
		f, err := os.Open(urlx.GetURLFilePath(&location))
		if err != nil {
			errs <- errors.WithStack(herodot.
				ErrInternalServerError.
				WithReasonf(
					`Unable to fetch JSON Web Keys from location "%s" because "%s".`,
					location.String(),
					err,
				),
			)
			return
		}
		defer f.Close()

		reader = f
	case "https":
		fallthrough
	case "http":
		res, err := s.client.Get(location.String())
		if err != nil {
			errs <- errors.WithStack(herodot.
				ErrInternalServerError.
				WithReasonf(
					`Unable to fetch JSON Web Keys from location "%s" because "%s".`,
					location.String(),
					err,
				),
			)
			return
		}
		defer res.Body.Close()

		if res.StatusCode < 200 || res.StatusCode >= 400 {
			errs <- errors.WithStack(herodot.
				ErrInternalServerError.
				WithReasonf(
					`Expected successful status code from location "%s", but received code "%d".`,
					location.String(),
					res.StatusCode,
				),
			)
			return
		}

		reader = res.Body
	default:
		errs <- errors.WithStack(herodot.
			ErrInternalServerError.
			WithReasonf(
				`Unable to fetch JSON Web Keys from location "%s" because URL scheme "%s" is not supported.`,
				location.String(),
				location.Scheme,
			),
		)
		return
	}

	var set jose.JSONWebKeySet
	if err := json.NewDecoder(reader).Decode(&set); err != nil {
		errs <- errors.WithStack(herodot.
			ErrInternalServerError.
			WithReasonf(
				`Unable to decode JSON Web Keys from location "%s" because "%s".`,
				location.String(),
				err,
			),
		)
		return
	}

	s.Lock()
	s.keys[location.String()] = set
	s.fetchedAt[location.String()] = time.Now().UTC()
	s.Unlock()
}
