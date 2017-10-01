package director

import (
	"testing"
	"net/http"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"regexp"
	"net/url"
)

func newTestRequest(method, u, token, ip string, forwardedIP []string) *http.Request {
	us, _ := url.Parse(u)
	return &http.Request{
		Method: method,
		URL:    us,
		Header: http.Header{
			"Authorization":   []string{"bearer " + token},
			"X-Forwarded-For": forwardedIP,
		},
		RemoteAddr: ip,
	}
}

func newMatcher(u string, methods ...string) *URLMatcher {
	c, _ := regexp.Compile(u)
	return &URLMatcher{
		Methods:  methods,
		URLMatch: c,
	}
}

func TestAccessRequestBuilder(t *testing.T) {
	for k, tc := range []struct {
		r           *http.Request
		rules       []Rule
		assert      func(t *testing.T, request *swagger.WardenTokenAccessRequest, err error)
		description string
	}{
		{
			description: "passes because rule matches",
			r:           newTestRequest("POST", "https://mydomain.com/users/1234", "some-token", "127.0.0.1", nil),
			rules: []Rule{
				{
					Action:  "get:$1", Resource: "mydomain:users:$1",
					Scopes:  []string{"users.get"},
					Matcher: newMatcher("mydomain.com/users/([0-9]+)", "PoSt"),
				},
			},
			assert: func(t *testing.T, request *swagger.WardenTokenAccessRequest, err error) {
				require.NoError(t, err)
				assert.Equal(t, "some-token", request.Token)
				assert.EqualValues(t, []string{"users.get"}, request.Scopes)
				assert.EqualValues(t, "mydomain:users:1234", request.Resource)
				assert.EqualValues(t, "get:1234", request.Action)
				assert.Equal(t, "127.0.0.1", request.Context["remoteIpAddress"])
			},
		},
		{
			description: "failse because rule is public",
			r:           newTestRequest("POST", "https://mydomain.com/users/1234", "some-token", "", nil),
			rules:       []Rule{{Matcher: newMatcher("mydomain.com/users/([0-9]+)", "PoSt"), Public: true}},
			assert: func(t *testing.T, request *swagger.WardenTokenAccessRequest, err error) {
				assert.EqualError(t, err, ErrPublicRule.Error())
				assert.Nil(t, request)
			},
		},
		{
			description: "fails because no bearer token is set",
			r:           newTestRequest("POST", "https://mydomain.com/users/1234", "", "", nil),
			rules:       []Rule{{Matcher: newMatcher("mydomain.com/users/([0-9]+)", "PoSt")}},
			assert: func(t *testing.T, request *swagger.WardenTokenAccessRequest, err error) {
				assert.EqualError(t, err, ErrMissingBearerToken.Error())
				assert.Nil(t, request)
			},
		},
	} {
		t.Run("case="+strconv.Itoa(k)+"/description="+tc.description, func(t *testing.T) {
			builder := &WardenRequestBuilder{CachedMatcher{Rules: tc.rules}}
			request, err := builder.BuildWardenRequest(tc.r)
			tc.assert(t, request, err)
		})
	}
}
