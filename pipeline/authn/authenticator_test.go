package authn_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ory/x/urlx"

	"github.com/ory/oathkeeper/pipeline/authn"
)

const (
	key = "key"
	val = "value"
)

func TestSetHeader(t *testing.T) {
	assert := assert.New(t)
	for k, tc := range []struct {
		a    *authn.AuthenticationSession
		desc string
	}{
		{
			a:    &authn.AuthenticationSession{},
			desc: "should initiate Header field if it is nil",
		},
		{
			a:    &authn.AuthenticationSession{Header: map[string][]string{}},
			desc: "should add a header to AuthenticationSession",
		},
	} {
		t.Run(fmt.Sprintf("case=%d/description=%s", k, tc.desc), func(t *testing.T) {
			tc.a.SetHeader(key, val)

			assert.NotNil(tc.a.Header)
			assert.Len(tc.a.Header, 1)
			assert.Equal(tc.a.Header.Get(key), val)
		})
	}
}

func TestCopy(t *testing.T) {
	assert := assert.New(t)
	original := &authn.AuthenticationSession{
		Subject: "ab",
		Extra:   map[string]interface{}{"a": "b", "b": map[string]string{"a:": "b"}},
		Header:  http.Header{"foo": {"bar", "baz"}},
		MatchContext: authn.MatchContext{
			RegexpCaptureGroups: []string{"a", "b"},
			URL:                 urlx.ParseOrPanic("https://foo/bar"),
		},
	}

	copied := original.Copy()
	copied.Subject = "ba"
	copied.Extra["baz"] = "bar"
	copied.Header.Add("bazbar", "bar")
	copied.MatchContext.URL.Host = "asdf"
	copied.MatchContext.RegexpCaptureGroups[0] = "b"

	assert.NotEqual(original.Subject, copied.Subject)
	assert.NotEqual(original.Extra, copied.Extra)
	assert.NotEqual(original.Header, copied.Header)
	assert.NotEqual(original.MatchContext.URL.Host, copied.MatchContext.URL.Host)
	assert.NotEqual(original.MatchContext.RegexpCaptureGroups, copied.MatchContext.RegexpCaptureGroups)
}
