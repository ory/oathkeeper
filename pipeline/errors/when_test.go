// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gobuffalo/httptest"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
)

func TestMatchesMIME(t *testing.T) {
	// application/json; charset=utf-8
	assert.False(t, matchesAcceptMIME("application/json", []string{"application/xml"}))
	assert.True(t, matchesAcceptMIME("application/json", []string{"application/xml", "application/json"}))

	assert.False(t, matchesAcceptMIME("application/*", []string{"application/json"}))
	assert.False(t, matchesAcceptMIME("*/*", []string{"application/json"}))

	assert.True(t, matchesAcceptMIME("application/*", []string{"application/*"}))
	assert.True(t, matchesAcceptMIME("*/*", []string{"*/*"}))

	assert.True(t, matchesAcceptMIME("text/html;q=0.9, application/xml;q=0.8, application/json;q=0.7", []string{"application/xml", "application/json"}))
	assert.True(t, matchesAcceptMIME("application/xml", []string{"application/xml", "application/json"}))

	jsonUTF8, err := clearMIMEParams("application/json; charset=utf-8")
	require.NoError(t, err)
	assert.Equal(t, "application/json", jsonUTF8)

	assert.True(t, matchesAcceptMIME(jsonUTF8, []string{"application/xml", "application/json"}))
	assert.True(t, matchesAcceptMIME(jsonUTF8, []string{"application/json"}))
	assert.True(t, matchesAcceptMIME(jsonUTF8, []string{"application/*"}))
	assert.True(t, matchesAcceptMIME(jsonUTF8, []string{"*/*"}))
}

func TestMatchesWhen(t *testing.T) {
	mixedAccept := func(t *testing.T, r *http.Request) { r.Header.Set("Accept", "application/json,text/html") }
	jsonAccept := func(t *testing.T, r *http.Request) { r.Header.Set("Accept", "application/json") }

	chromeAccept := func(t *testing.T, r *http.Request) {
		r.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	}
	firefoxAccept := func(t *testing.T, r *http.Request) {
		r.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	}

	jsonContentType := func(t *testing.T, r *http.Request) { r.Header.Set("Content-Type", "application/json") }
	withIPs := func(remote string, forwarded ...string) func(t *testing.T, r *http.Request) {
		return func(t *testing.T, r *http.Request) {
			r.RemoteAddr = remote
			if len(forwarded) > 0 {
				r.Header.Set("X-Forwarded-For", strings.Join(forwarded, ", "))
			}
		}
	}
	combine := func(fs ...func(t *testing.T, r *http.Request)) func(t *testing.T, r *http.Request) {
		return func(t *testing.T, r *http.Request) {
			for _, f := range fs {
				f(t, r)
			}
		}
	}

	for k, tc := range []struct {
		ee   error
		in   error
		w    Whens
		init func(t *testing.T, r *http.Request)
	}{
		{in: errors.New("err")},
		{
			w:  Whens{When{Error: []string{statusText(http.StatusForbidden)}}},
			in: &herodot.ErrNotFound,
			ee: ErrDoesNotMatchWhen,
		},
		{
			w:  Whens{When{Error: []string{"forbidden"}}},
			in: errors.WithStack(&herodot.ErrNotFound),
			ee: ErrDoesNotMatchWhen,
		},
		{
			w: Whens{
				When{Error: []string{statusText(http.StatusForbidden), statusText(http.StatusNotFound)}},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
		},
		{
			w: Whens{
				When{Error: []string{statusText(http.StatusForbidden)}},
				When{Error: []string{statusText(http.StatusForbidden), statusText(http.StatusNotFound)}},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
		},
		{
			w: Whens{
				When{
					Error:   []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
		},
		{
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							ContentType: []string{"application/json"},
						},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
			ee: ErrDoesNotMatchWhen,
		},
		{
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							Accept: []string{"application/json"},
						},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
			ee: ErrDoesNotMatchWhen,
		},
		{
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						RemoteIP: &WhenRequestRemoteIP{Match: []string{"192.168.1.0/24"}},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
			ee: ErrDoesNotMatchWhen,
		},
		{
			init: combine(jsonAccept),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							Accept: []string{"application/xml"},
						},
					},
				},
				When{
					Error: []string{statusText(http.StatusUnauthorized)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							Accept: []string{"application/json"},
						},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
			ee: ErrDoesNotMatchWhen,
		},
		{
			init: combine(jsonAccept),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							Accept: []string{"application/xml"},
						},
					},
				},
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							Accept: []string{"application/json"},
						},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
		},
		{
			init: combine(jsonAccept, jsonContentType),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							Accept: []string{"application/xml"},
						},
					},
				},
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							Accept:      []string{"application/json"},
							ContentType: []string{"application/xml"},
						},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
			ee: ErrDoesNotMatchWhen,
		},
		{
			init: combine(jsonAccept, jsonContentType),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							ContentType: []string{"application/xml"},
						},
					},
				},
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							Accept:      []string{"application/json"},
							ContentType: []string{"application/json"},
						},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
		},
		{
			init: combine(jsonAccept, jsonContentType),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							ContentType: []string{"application/xml"},
						},
					},
				},
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						RemoteIP: &WhenRequestRemoteIP{Match: []string{"192.168.1.0/24"}},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
			ee: ErrDoesNotMatchWhen,
		},
		{
			init: combine(jsonAccept, jsonContentType, withIPs("192.168.1.1:1234")),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							ContentType: []string{"application/xml"},
						},
					},
				},
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						RemoteIP: &WhenRequestRemoteIP{Match: []string{"192.168.1.0/24"}},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
		},
		{
			init: combine(jsonAccept, jsonContentType, withIPs("127.0.0.1:123")),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						RemoteIP: &WhenRequestRemoteIP{Match: []string{"192.168.1.0/24"}},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
			ee: ErrDoesNotMatchWhen,
		},
		{
			init: combine(jsonAccept, jsonContentType, withIPs("127.0.0.1:123", "127.0.0.2")),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						RemoteIP: &WhenRequestRemoteIP{Match: []string{"192.168.1.0/24"}},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
			ee: ErrDoesNotMatchWhen,
		},
		{
			init: combine(jsonAccept, jsonContentType, withIPs("127.0.0.1:123", "127.0.0.2", "127.0.0.3")),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						RemoteIP: &WhenRequestRemoteIP{Match: []string{"192.168.1.0/24"}},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
			ee: ErrDoesNotMatchWhen,
		},
		{
			init: combine(jsonAccept, jsonContentType, withIPs("127.0.0.1:123", "127.0.0.2", "127.0.0.3", "192.168.1.2")),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						RemoteIP: &WhenRequestRemoteIP{Match: []string{"192.168.1.0/24"}},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
			ee: ErrDoesNotMatchWhen,
		},
		{
			init: combine(jsonAccept, jsonContentType, withIPs("127.0.0.1:123", "127.0.0.2", "127.0.0.3", "192.168.1.2")),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						RemoteIP: &WhenRequestRemoteIP{RespectForwardedForHeader: true, Match: []string{"192.168.1.0/24"}},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
		},
		{
			init: combine(mixedAccept, jsonContentType, withIPs("127.0.0.1:123", "127.0.0.2", "127.0.0.3", "192.168.1.2")),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						RemoteIP: &WhenRequestRemoteIP{RespectForwardedForHeader: true, Match: []string{"192.168.1.0/24"}},
						Header: &WhenRequestHeader{
							ContentType: []string{"application/json"},
							Accept:      []string{"application/xml", "application/json"},
						},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
		},
		{
			init: combine(mixedAccept, jsonContentType),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							ContentType: []string{"application/*"},
							Accept:      []string{"text/html"},
						},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
		},
		{
			init: combine(mixedAccept, jsonContentType),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							ContentType: []string{"application/json"},
							Accept:      []string{"text/*"},
						},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
		},
		{
			init: combine(chromeAccept),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							Accept: []string{"application/json"},
						},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
			ee: ErrDoesNotMatchWhen,
		},
		{
			init: combine(firefoxAccept),
			w: Whens{
				When{
					Error: []string{statusText(http.StatusNotFound)},
					Request: &WhenRequest{
						Header: &WhenRequestHeader{
							Accept: []string{"application/json"},
						},
					},
				},
			},
			in: errors.WithStack(&herodot.ErrNotFound),
			ee: ErrDoesNotMatchWhen,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			r := httptest.NewRequest("GET", "/test", nil)
			if tc.init != nil {
				tc.init(t, r)
			}

			err := MatchesWhen(tc.w, r, tc.in)
			if tc.ee != nil {
				require.EqualError(t, err, tc.ee.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestStatusText(t *testing.T) {
	assert.Equal(t, "not_found", statusText(http.StatusNotFound))
	assert.Equal(t, "im_a_teapot", statusText(http.StatusTeapot))
}
