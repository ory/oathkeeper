package credentials

import (
	"context"
	"encoding/json"
	"github.com/ory/herodot"
	"github.com/ory/x/urlx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

var sets = [...]json.RawMessage{
	json.RawMessage(`{"keys":[{"kty":"oct","kid":"c61308cc-faef-4b98-99c3-839f513ac296","k":"I2_YrZxll-Uq65GKjnJq4u7uNub8hG5cBvlHRz03w94","alg":"HS256"}]}`),
	json.RawMessage(`{"keys":[{"kty":"oct","kid":"2aeaef79-7233-4a59-95bf-e32151d3544b","k":"NJACtF9Hbivq3Q67LDtS_mbD33PHTTSlq7By7Wbm9tg","alg":"HS256"}]}`),
	json.RawMessage(`{"keys":[{"kty":"oct","kid":"392e1a6b-6ae1-48b8-bea3-2fe09447805c","k":"Wp6sSiCjQQOp-bg7fifclpTpA2xrOujM7PYgP97_Sxg","alg":"HS256"},{"kty":"oct","kid":"8e884167-1300-4f58-8cc1-81af68f878a8","k":"oX3Vu6g_ezpwFK19EAiElxFLOLHf0R8i35WoAUQUU5w","alg":"HS256"}]}`),
	json.RawMessage(`invalid json ¯\_(ツ)_/¯`),
}

func TestNewJWKSFetcherStrategy(t *testing.T) {
	const maxWait = time.Millisecond * 100

	l := logrus.New()
	l.Level = logrus.DebugLevel

	w := herodot.NewJSONWriter(l)
	s := NewJWKSFetcherStrategyWithTimeout(l, maxWait, maxWait*3, maxWait*7)

	timeOutServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		time.Sleep(maxWait * 2)
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

	invalidServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.Write(sets[3])
	}))
	defer invalidServer.Close()

	uris := []url.URL{
		*urlx.ParseOrPanic(timeOutServer.URL),
		*urlx.ParseOrPanic(slowServer.URL),
		*urlx.ParseOrPanic(fastServer.URL),
		*urlx.ParseOrPanic(invalidServer.URL),
	}

	t.Run("name=should result in error because server times out", func(t *testing.T) {
		_, err := s.Resolve(context.Background(), uris, "c61308cc-faef-4b98-99c3-839f513ac296")
		require.Error(t, err)
	})

	t.Run("name=should result in error because key id does not exist", func(t *testing.T) {
		_, err := s.Resolve(context.Background(), uris, "i-do-not-exist")
		require.Error(t, err)
	})

	t.Run("name=should find the key even if the server is slow", func(t *testing.T) {
		key, err := s.Resolve(context.Background(), uris, "2aeaef79-7233-4a59-95bf-e32151d3544b")
		require.NoError(t, err)
		assert.Equal(t, "2aeaef79-7233-4a59-95bf-e32151d3544b", key.KeyID)
	})

	t.Run("name=should find the key when the server works normally and when it sends multiple keys", func(t *testing.T) {
		key, err := s.Resolve(context.Background(), uris, "392e1a6b-6ae1-48b8-bea3-2fe09447805c")
		require.NoError(t, err)
		assert.Equal(t, "392e1a6b-6ae1-48b8-bea3-2fe09447805c", key.KeyID)

		key, err = s.Resolve(context.Background(), uris, "8e884167-1300-4f58-8cc1-81af68f878a8")
		require.NoError(t, err)
		assert.Equal(t, "8e884167-1300-4f58-8cc1-81af68f878a8", key.KeyID)
	})

	t.Run("name=should find the previously timed out key because enough time has passed", func(t *testing.T) {
		key, err := s.Resolve(context.Background(), uris, "c61308cc-faef-4b98-99c3-839f513ac296")
		require.NoError(t, err)
		assert.Equal(t, "c61308cc-faef-4b98-99c3-839f513ac296", key.KeyID)
	})

	t.Run("name=should find the key even if the upstream server is no longer active", func(t *testing.T) {
		fastServer.Close()
		key, err := s.Resolve(context.Background(), uris, "392e1a6b-6ae1-48b8-bea3-2fe09447805c")
		require.NoError(t, err)
		assert.Equal(t, "392e1a6b-6ae1-48b8-bea3-2fe09447805c", key.KeyID)
	})

	time.Sleep(maxWait)

	t.Run("name=should no longer find the key if the remote does not find it", func(t *testing.T) {
		key, err := s.Resolve(context.Background(), uris, "392e1a6b-6ae1-48b8-bea3-2fe09447805c")
		require.NoError(t, err)
		assert.Equal(t, "392e1a6b-6ae1-48b8-bea3-2fe09447805c", key.KeyID)
	})
}
