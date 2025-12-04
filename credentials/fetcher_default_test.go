// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package credentials

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"

	"github.com/ory/x/logrusx"

	"github.com/ory/oathkeeper/x"

	"github.com/ory/herodot"
	"github.com/ory/x/urlx"

	"github.com/ory/oathkeeper/internal/cloudstorage"
)

var sets = [...]json.RawMessage{
	json.RawMessage(`{"keys":[{"use":"sig","kty":"oct","kid":"c61308cc-faef-4b98-99c3-839f513ac296","k":"I2_YrZxll-Uq65GKjnJq4u7uNub8hG5cBvlHRz03w94","alg":"HS256"}]}`),
	json.RawMessage(`{"keys":[{"use":"sig","kty":"oct","kid":"2aeaef79-7233-4a59-95bf-e32151d3544b","k":"NJACtF9Hbivq3Q67LDtS_mbD33PHTTSlq7By7Wbm9tg","alg":"HS256"}]}`),
	json.RawMessage(`{"keys":[{"use":"sig","kty":"oct","kid":"392e1a6b-6ae1-48b8-bea3-2fe09447805c","k":"Wp6sSiCjQQOp-bg7fifclpTpA2xrOujM7PYgP97_Sxg","alg":"HS256"},{"use":"sig","kty":"oct","kid":"8e884167-1300-4f58-8cc1-81af68f878a8","k":"oX3Vu6g_ezpwFK19EAiElxFLOLHf0R8i35WoAUQUU5w","alg":"HS256"}]}`),
	json.RawMessage(`invalid json ¯\_(ツ)_/¯`),
}

type reg struct{}

func (*reg) Logger() *logrusx.Logger {
	return logrusx.New("", "", logrusx.ForceLevel(logrus.DebugLevel))
}

func (*reg) Tracer() trace.Tracer {
	return nooptrace.NewTracerProvider().Tracer("")
}

func TestFetcherDefault(t *testing.T) {
	t.Parallel()

	const maxWait = time.Millisecond * 100
	const JWKsTTL = maxWait * 7
	const timeoutServerDelay = maxWait * 2

	l := logrusx.New("", "", logrusx.ForceLevel(logrus.DebugLevel))
	w := herodot.NewJSONWriter(l)
	s := NewFetcherDefault(&reg{}, maxWait, JWKsTTL)

	timeOutServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		time.Sleep(timeoutServerDelay)
		w.Write(rw, r, sets[0])
	}))
	defer timeOutServer.Close()

	slowServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		time.Sleep(maxWait / 2)
		w.Write(rw, r, sets[1])
	}))
	defer slowServer.Close()

	fastServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		w.Write(rw, r, sets[2])
	}))
	defer fastServer.Close()

	invalidServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write(sets[3])
	}))
	defer invalidServer.Close()

	uris := []url.URL{
		*x.ParseURLOrPanic(timeOutServer.URL),
		*x.ParseURLOrPanic(slowServer.URL),
		*x.ParseURLOrPanic(fastServer.URL),
		*x.ParseURLOrPanic(invalidServer.URL),
		*x.ParseURLOrPanic("file://../test/stub/jwks-hs.json"),
		*x.ParseURLOrPanic("file://../test/stub/jwks-rsa-single.json"),
		*x.ParseURLOrPanic("file://../test/stub/jwks-rsa-multiple.json"),
	}

	t.Run("name=should result in error because server times out", func(t *testing.T) {
		_, err := s.ResolveKey(context.Background(), uris, "c61308cc-faef-4b98-99c3-839f513ac296", "sig")
		require.Error(t, err)
	})

	t.Run("name=should result in error because key id does not exist", func(t *testing.T) {
		_, err := s.ResolveKey(context.Background(), uris, "i-do-not-exist", "sig")
		require.Error(t, err)
	})

	t.Run("name=should find the key even if the server is slow", func(t *testing.T) {
		key, err := s.ResolveKey(context.Background(), uris, "2aeaef79-7233-4a59-95bf-e32151d3544b", "sig")
		require.NoError(t, err)
		assert.Equal(t, "2aeaef79-7233-4a59-95bf-e32151d3544b", key.KeyID)
	})

	t.Run("name=should find the key when the server works normally and when it sends multiple keys", func(t *testing.T) {
		key, err := s.ResolveKey(context.Background(), uris, "392e1a6b-6ae1-48b8-bea3-2fe09447805c", "sig")
		require.NoError(t, err)
		assert.Equal(t, "392e1a6b-6ae1-48b8-bea3-2fe09447805c", key.KeyID)

		key, err = s.ResolveKey(context.Background(), uris, "8e884167-1300-4f58-8cc1-81af68f878a8", "sig")
		require.NoError(t, err)
		assert.Equal(t, "8e884167-1300-4f58-8cc1-81af68f878a8", key.KeyID)
	})

	t.Run("name=should find the previously timed out key because enough time has passed", func(t *testing.T) {
		key, err := s.ResolveKey(context.Background(), uris, "c61308cc-faef-4b98-99c3-839f513ac296", "sig")
		require.NoError(t, err)
		assert.Equal(t, "c61308cc-faef-4b98-99c3-839f513ac296", key.KeyID)
	})

	t.Run("name=should find the key even if the upstream server is no longer active", func(t *testing.T) {
		t.Parallel()
		fastServer.Close()
		key, err := s.ResolveKey(context.Background(), uris, "392e1a6b-6ae1-48b8-bea3-2fe09447805c", "sig")
		require.NoError(t, err)
		assert.Equal(t, "392e1a6b-6ae1-48b8-bea3-2fe09447805c", key.KeyID)

		t.Run("name=should no longer find the key if the remote does not find it", func(t *testing.T) {
			time.Sleep(maxWait)

			key, err := s.ResolveKey(context.Background(), uris, "392e1a6b-6ae1-48b8-bea3-2fe09447805c", "sig")
			require.NoError(t, err)
			assert.Equal(t, "392e1a6b-6ae1-48b8-bea3-2fe09447805c", key.KeyID)
		})
	})

	t.Run("name=should fetch keys from the file system", func(t *testing.T) {
		t.Parallel()
		key, err := s.ResolveKey(context.Background(), uris, "81be3441-5303-4c52-b00d-bbdfadc75633", "sig")
		require.NoError(t, err)
		assert.Equal(t, "81be3441-5303-4c52-b00d-bbdfadc75633", key.KeyID)

		key, err = s.ResolveKey(context.Background(), uris, "3e0edde4-12ad-425d-a783-135f46eac57e", "sig")
		require.NoError(t, err)
		assert.Equal(t, "3e0edde4-12ad-425d-a783-135f46eac57e", key.KeyID)

		key, err = s.ResolveKey(context.Background(), uris, "f4190122-ae96-4c29-8b79-56024e459d80", "sig")
		require.NoError(t, err)
		assert.Equal(t, "f4190122-ae96-4c29-8b79-56024e459d80", key.KeyID)
	})

	t.Run("name=should resolve all the json web key sets", func(t *testing.T) {
		sets, err := s.ResolveSets(context.Background(), uris)
		require.NoError(t, err)
		assert.Len(t, sets, len(uris)-1) // this is -1 because on url is invalid!

		var check = func(kid string) (found bool) {
			for _, set := range sets {
				if len(set.Key(kid)) > 0 {
					found = true
					break
				}
			}
			return
		}

		// Check if some random keys exists
		assert.True(t, check("f4190122-ae96-4c29-8b79-56024e459d80"))
		assert.True(t, check("8e884167-1300-4f58-8cc1-81af68f878a8"))
	})

	time.Sleep(
		timeoutServerDelay +
			JWKsTTL +
			(time.Millisecond * 100)) // wait so the fetched key reaches ttl
	// change "alg" for "c61308cc-faef-4b98-99c3-839f513ac296",
	// so we are sure we get the "stale" data in `name=should find the previously fetched key if the refresh request times out`
	sets[0] = json.RawMessage(`{"keys":[{"use":"sig","kty":"oct","kid":"c61308cc-faef-4b98-99c3-839f513ac296","k":"I2_YrZxll-Uq65GKjnJq4u7uNub8hG5cBvlHRz03w94","alg":"RS256"}]}`)

	t.Run("name=should find the previously fetched key if the refresh request times out", func(t *testing.T) {
		key, err := s.ResolveKey(context.Background(), uris, "c61308cc-faef-4b98-99c3-839f513ac296", "sig")
		require.NoError(t, err)
		assert.Equal(t, "HS256", key.Algorithm)
	})

	t.Run("name=should fetch from s3 object storage", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		s := NewFetcherDefault(&reg{}, maxWait, JWKsTTL, WithURLMux(cloudstorage.NewTestURLMux(t)))

		key, err := s.ResolveKey(ctx, []url.URL{
			*urlx.ParseOrPanic("s3://oathkeeper-test-bucket/path/prefix/jwks.json"),
		}, "81be3441-5303-4c52-b00d-bbdfadc75633", "sig")
		require.NoError(t, err)
		assert.Equal(t, "81be3441-5303-4c52-b00d-bbdfadc75633", key.KeyID)
	})

	t.Run("name=should fetch from gs object storage", func(t *testing.T) {
		t.Parallel()
		t.Skip("golden file needs to be updated")
		ctx := context.Background()
		s := NewFetcherDefault(&reg{}, maxWait, JWKsTTL, WithURLMux(cloudstorage.NewTestURLMux(t)))

		key, err := s.ResolveKey(ctx, []url.URL{
			*urlx.ParseOrPanic("gs://oathkeeper-test-bucket/path/prefix/jwks.json"),
		}, "81be3441-5303-4c52-b00d-bbdfadc75633", "sig")
		require.NoError(t, err)
		assert.Equal(t, "81be3441-5303-4c52-b00d-bbdfadc75633", key.KeyID)
	})

	t.Run("name=should fetch from azure object storage", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		s := NewFetcherDefault(&reg{}, maxWait, JWKsTTL, WithURLMux(cloudstorage.NewTestURLMux(t)))

		jwkKey, err := s.ResolveKey(ctx, []url.URL{
			*urlx.ParseOrPanic("azblob://path/prefix/jwks.json"),
		}, "81be3441-5303-4c52-b00d-bbdfadc75633", "sig")
		require.NoError(t, err)
		assert.Equal(t, "81be3441-5303-4c52-b00d-bbdfadc75633", jwkKey.KeyID)
	})
}
