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
	"github.com/ory/herodot"
	"github.com/ory/x/httpx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/square/go-jose.v2"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type reasoner interface {
	Reason() string
}

type JWKSFetcherStrategy struct {
	sync.RWMutex

	ttl         time.Duration
	cancelAfter time.Duration
	client      *http.Client
	keys        map[string]jose.JSONWebKeySet
	fetchedAt   map[string]time.Time
	l           logrus.FieldLogger
}

func NewJWKSFetcherStrategy(l logrus.FieldLogger) *JWKSFetcherStrategy {
	return NewJWKSFetcherStrategyWithTimeout(l, time.Second, time.Second*10, time.Minute*5)
}

// NewJWKSFetcherStrategyWithTimeout returns a new JWKS Fetcher with:
//
// - cancelAfter: If reached, the fetcher will stop waiting for responses and return an error.
// - waitForResponse: While the fetcher might stop waiting for responses, we will give the server more time to respond
//		and add the keys to the registry unless waitForResponse is reached in which case we'll terminate the request.
func NewJWKSFetcherStrategyWithTimeout(l logrus.FieldLogger, cancelAfter time.Duration, waitForResponse time.Duration, ttl time.Duration) *JWKSFetcherStrategy {
	return &JWKSFetcherStrategy{
		cancelAfter: cancelAfter,
		l:           l,
		ttl:         ttl,
		keys:        make(map[string]jose.JSONWebKeySet),
		fetchedAt:   make(map[string]time.Time),
		client: &http.Client{
			Transport: httpx.NewDefaultResilientRoundTripper(cancelAfter/2, waitForResponse),
		},
	}
}

func (s *JWKSFetcherStrategy) Resolve(ctx context.Context, locations []url.URL, kid string) (*jose.JSONWebKey, error) {
	if key := s.key(kid, locations); key != nil {
		return key, nil
	}

	ctx, _ = context.WithTimeout(ctx, s.cancelAfter)
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
		s.l.Errorf("Ignoring JSON Web Keys from at least one URI because the request timed out waiting for a response.")
	case <-done:
		// We're done!
	}

	if key := s.key(kid, locations); key != nil {
		return key, nil
	}

	return nil, errors.WithStack(herodot.ErrForbidden.WithReason("The provided JSON Web Token was signed with an unknown key.").WithDebug("Check that the provided JSON Web Key URIs contain a key that can verify the signature of the provided JSON Web Token."))
}

func (s *JWKSFetcherStrategy) key(kid string, locations []url.URL) *jose.JSONWebKey {
	for _, l := range locations {
		s.RLock()
		keys, ok1 := s.keys[l.String()]
		fetchedAt, ok2 := s.fetchedAt[l.String()]
		s.RUnlock()

		if !ok1 || !ok2 || fetchedAt.Add(s.ttl).Before(time.Now().UTC()) {
			continue
		}

		if kk := keys.Key(kid); len(kk) > 0 {
			return &kk[0]
		}
	}

	return nil
}

func (s *JWKSFetcherStrategy) resolveAll(done chan struct{}, errs chan error, locations []url.URL) {
	var wg sync.WaitGroup

	for _, l := range locations {
		wg.Add(1)
		go s.resolve(&wg, errs, l)
	}

	wg.Wait()
	close(done)
	close(errs)
}

func (s *JWKSFetcherStrategy) resolve(wg *sync.WaitGroup, errs chan error, location url.URL) {
	defer wg.Done()

	res, err := s.client.Get(location.String())
	if err != nil {
		errs <- errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`Unable to fetch JSON Web Keys from location "%s" because %s".`, location.String(), err))
		return
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		errs <- errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`Expected successful status code from location "%s", but received code "%d".`, location.String(), res.StatusCode))
		return
	}

	var set jose.JSONWebKeySet
	if err := json.NewDecoder(res.Body).Decode(&set); err != nil {
		errs <- errors.WithStack(herodot.ErrInternalServerError.WithReasonf(`Unable to decode JSON Web Keys from location "%s" because "%s".`, location.String(), err))
		return
	}

	s.Lock()
	s.keys[location.String()] = set
	s.fetchedAt[location.String()] = time.Now().UTC()
	s.Unlock()
}
