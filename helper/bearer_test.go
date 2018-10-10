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

package helper

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBearerToken(t *testing.T) {
	r := httptest.NewRequest("GET", "https://a.b.c/xyz", nil)
	token := BearerTokenFromRequest(r)
	assert.Equal(t, "", token)

	r.Header.Add("Authorization", "Bearer |token|")
	token = BearerTokenFromRequest(r)
	assert.Equal(t, "|token|", token)

	u, _ := url.Parse("https://a.b.c/xyz")
	e := u.Query()
	e.Add("access_token", "|token|")
	u.RawQuery = e.Encode()

	r = httptest.NewRequest("GET", u.RequestURI(), nil)
	token = BearerTokenFromRequest(r)
	assert.Equal(t, "|token|", token)

	body := strings.NewReader(e.Encode())
	r = httptest.NewRequest("POST", "https://a.b.c/xyz", body)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	token = BearerTokenFromRequest(r)
	assert.Equal(t, "|token|", token)
}

func TestBearerTokenFromCookie(t *testing.T) {
	//configCookies map[string]CookieTokenConfig, r *http.Request) (string, error) {
	confMap := map[string]CookieTokenConfig{
		"_cookie1": {},
		"_cookie2": {
			ValueFormat: "rawstd-base64-json",
			TokenKey:    "token",
		},
		"_cookie3": {
			ValueFormat: "garbled",
			TokenKey:    "absent",
		},
		"_cookie4": {
			ValueFormat: "",
			TokenKey:    "absent",
		},
		"_cookie5": {
			ValueFormat: "rawurl-base64-json",
		},
	}
	r := httptest.NewRequest("POST", "https://a.b.c/xyz", nil)
	r.AddCookie(&http.Cookie{
		Name:  "irrelevant",
		Value: "123",
	})
	token, err := BearerTokenFromCookie(confMap, r)
	assert.Empty(t, token)
	assert.NoError(t, err)

	jazon := []byte(`{"access_token": "123456789","other": "abcd"}`)
	value := base64.RawStdEncoding.EncodeToString(jazon)
	r.AddCookie(&http.Cookie{
		Name:  "_cookie1",
		Value: value,
	})
	token, err = BearerTokenFromCookie(confMap, r)
	assert.NoError(t, err)
	assert.Equal(t, "123456789", token)

	r = httptest.NewRequest("POST", "https://a.b.c/xyz", nil)
	jazon = []byte(`{"token": "123456789","other": "abcd"}`)
	value = base64.RawStdEncoding.EncodeToString(jazon)
	r.AddCookie(&http.Cookie{
		Name:  "_cookie2",
		Value: value,
	})
	token, err = BearerTokenFromCookie(confMap, r)
	assert.NoError(t, err)
	assert.Equal(t, "123456789", token)

	r = httptest.NewRequest("POST", "https://a.b.c/xyz", nil)
	jazon = []byte(`{"token": "123456789","other": "abcd"}`)
	value = base64.RawStdEncoding.EncodeToString(jazon)
	r.AddCookie(&http.Cookie{
		Name:  "_cookie3",
		Value: value,
	})
	token, err = BearerTokenFromCookie(confMap, r)
	assert.Empty(t, token)
	assert.Error(t, err)

	r = httptest.NewRequest("POST", "https://a.b.c/xyz", nil)
	jazon = []byte(`{"token": "123456789","other": "abcd"}`)
	value = base64.RawStdEncoding.EncodeToString(jazon)
	r.AddCookie(&http.Cookie{
		Name:  "_cookie4",
		Value: value,
	})
	token, err = BearerTokenFromCookie(confMap, r)
	assert.Empty(t, token)
	assert.NoError(t, err)

	r = httptest.NewRequest("POST", "https://a.b.c/xyz", nil)
	jazon = []byte(`{"access_token": "123456789","other": "abcd"}`)
	value = base64.RawURLEncoding.EncodeToString(jazon)
	r.AddCookie(&http.Cookie{
		Name:  "_cookie5",
		Value: value,
	})
	token, err = BearerTokenFromCookie(confMap, r)
	assert.NoError(t, err)
	assert.Equal(t, "123456789", token)
}
