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
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/herodot"
	"github.com/ory/hive-cloud/hive-api/identity"
	"github.com/ory/hive-cloud/hive-api/session"
	"github.com/ory/oathkeeper/helper"
)

func TestIsAPIRequest(t *testing.T) {
	for k, tc := range []struct {
		h string
		e bool
	}{
		{
			e: false,
			h: `Host: prometheus.io
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:64.0) Gecko/20100101 Firefox/64.0
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
Referer: https://www.google.com/
DNT: 1
Connection: keep-alive
Cookie: foo=bar
Upgrade-Insecure-Requests: 1
Cache-Control: max-age=0`,
		},
		{
			e: true,
			h: `Host: ogs.google.com
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:64.0) Gecko/20100101 Firefox/64.0
Accept: */*
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
Referer: https://www.google.com/
Content-Type: application/x-www-form-urlencoded;charset=utf-8
Content-Length: 103
Origin: https://www.google.com
DNT: 1
Connection: keep-alive
Cookie: foo=bar
Cache-Control: max-age=0`,
		},
		{
			e: true,
			h: `Host: backoffice.oryapis.sh
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:64.0) Gecko/20100101 Firefox/64.0
Accept: */*
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
Referer: https://console.ory.sh/
origin: https://console.ory.sh
DNT: 1
Connection: keep-alive`,
		},
		{
			e: true,
			h: `Host: backoffice.oryapis.sh
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:64.0) Gecko/20100101 Firefox/64.0
Accept: */*
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
Referer: https://console.ory.sh/
origin: https://console.ory.sh
DNT: 1
Connection: keep-alive`,
		},
		{
			e: false,
			h: `Host: console.ory.sh
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:64.0) Gecko/20100101 Firefox/64.0
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate, br
DNT: 1
Connection: keep-alive
Upgrade-Insecure-Requests: 1`,
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			headers := http.Header{}
			for _, l := range strings.Split(tc.h, "\n") {
				ll := strings.Split(l, ":")
				headers.Set(ll[0], strings.TrimSpace(ll[1]))
			}

			assert.Equal(t, tc.e, isAPIRequest(&http.Request{
				Header: headers,
			}))
		})
	}
}

func TestAuthenticatorHive(t *testing.T) {
	assert.NotNil(t, NewAuthenticatorHive(nil, "", ""))
	assert.NotEmpty(t, NewAuthenticatorHive(nil, "", "").GetID())

	writer := herodot.NewJSONWriter(logrus.New())
	var sessionHandler func(w http.ResponseWriter, r *http.Request, _ httprouter.Params)

	h := httprouter.New()
	h.GET("/session", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		sessionHandler(w, r, ps)
	})
	h.GET("/init", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = fmt.Fprint(w, "login")
	})
	ts := httptest.NewServer(h)
	defer ts.Close()
	a := NewAuthenticatorHive(ts.Client(), ts.URL+"/session", ts.URL+"/init")

	t.Run("sub=findSession", func(t *testing.T) {
		for k, tc := range []struct {
			sh            func(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
			cv            string
			expectErr     error
			expectSession *session.Session
		}{
			{
				sh: func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					w.WriteHeader(http.StatusInternalServerError)
				},
				cv:        "foobar",
				expectErr: helper.ErrServerError,
			},
			{
				sh: func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					w.WriteHeader(http.StatusUnauthorized)
				},
				cv:        "foobar",
				expectErr: helper.ErrUnauthorized,
			},
			{
				sh: func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					writer.Write(w, r, &session.Session{
						SID:      "1-session-id",
						Metadata: json.RawMessage(`{}`),
					})
				},
				cv: "1-session-id",
				expectSession: &session.Session{
					SID:      "1-session-id",
					Metadata: json.RawMessage(`{}`),
				},
			},
			{
				sh: func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					// This request should not be called because we've cached it already. We check if the request
					// is being called again.
					panic("This should not have been called")
				},
				cv: "1-session-id",
				expectSession: &session.Session{
					SID:      "1-session-id",
					Metadata: json.RawMessage(`{}`),
				},
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				sessionHandler = tc.sh
				s, err := a.findSession(&http.Cookie{Value: tc.cv})
				if tc.expectErr != nil {
					require.EqualError(t, tc.expectErr, err.Error())
				} else {
					require.NoError(t, err)
					assert.EqualValues(t, tc.expectSession, s)
				}
			})
		}
	})

	t.Run("sub=Authenticate", func(t *testing.T) {
		for k, tc := range []struct {
			d          string
			sh         func(w http.ResponseWriter, r *http.Request, _ httprouter.Params)
			cookie     *http.Cookie
			config     json.RawMessage
			expectErr  error
			expectSess *AuthenticationSession
		}{
			{
				d:      "returns an authentication session when the session is found in hive",
				cookie: &http.Cookie{Value: "2-session-id", Name: session.DefaultSessionCookieName},
				sh: func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					writer.Write(w, r, &session.Session{
						SID: "2-session-id",
						Identity: &identity.Identity{
							URN: "some-subject",
						},
					})
				},
				expectSess: &AuthenticationSession{
					Subject: "some-subject",
				},
			},
			{
				d:      "when no cookie is set, return an unauthorized error",
				cookie: &http.Cookie{Value: "not-a-session", Name: "some-other-cookie"},
				sh: func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
					panic("should not have been called")
				},
				expectErr: helper.ErrForceResponse,
			},
			{
				d:         "should reject invalid on_unauthorized",
				config:    json.RawMessage(`{"on_unauthorized": "invalid"}`),
				expectErr: helper.ErrRuleMisconfiguration,
			},
		} {
			t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
				sessionHandler = tc.sh
				r := &http.Request{Header: http.Header{}}
				if tc.cookie != nil {
					r.AddCookie(tc.cookie)
				}

				s, err := a.Authenticate(r, json.RawMessage(tc.config), nil)

				if tc.expectErr != nil {
					require.EqualError(t, tc.expectErr, err.Error())
				} else {
					require.NoError(t, err, "%#v", err)
				}

				if tc.expectSess != nil {
					assert.Equal(t, tc.expectSess.Subject, s.Subject)
				}
			})
		}
	})
}
