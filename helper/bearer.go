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
	"encoding/json"
	"mime"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

const accessToken = "access_token"

// ErrInvalidCookieFormat indicates a malformed configuration
var ErrInvalidCookieFormat = errors.New("invalid cookie format specified. Supported values are: 'rawstd-base64-json' (default), 'rawurl-base64-json'")

// BearerTokenFromRequest retrieves an access token from request according
// to RFC6750
func BearerTokenFromRequest(r *http.Request) (token string) {
	auth := r.Header.Get("Authorization")
	split := strings.SplitN(auth, " ", 2)
	if len(split) == 2 && strings.EqualFold(split[0], "bearer") {
		token = split[1]
	}

	if token == "" && r.URL != nil {
		qs := r.URL.Query()
		token = qs.Get(accessToken)
	}

	if token == "" {
		ct := r.Header.Get("Content-Type")
		if ct != "" {
			contentType, _, _ := mime.ParseMediaType(ct)
			if contentType == "application/x-www-form-urlencoded" || contentType == "multipart/form-data" {
				token = r.FormValue(accessToken)
			}
		}
	}
	return
}

// CookieTokenConfig configures how a token may be stored in a cookie.
type CookieTokenConfig struct {
	// ValueFormat indicates how the value is parsed. At the moment, we assume base64 encoded JSON.
	ValueFormat string `json:"value_format"`

	// TokenKey indicates the key of the token value in the decoded cookie value. By default, we use key accessToken
	TokenKey string `json:"token_key"`
}

// BearerTokenFromCookie extracts a bearer token (e.g. access token, JWT) from a cookie container
func BearerTokenFromCookie(configCookies map[string]CookieTokenConfig, r *http.Request) (string, error) {
	// we have configured cookies: let's check if a token is contained in one of the recognized cookies.
	// We assume only one such cookie arrives. If several are in the request, any one containing a token
	// may be selected, in any order. Cookies with no token are skipped.
	var token string

	for k, v := range configCookies {
		ck, err := r.Cookie(k)
		if err == http.ErrNoCookie {
			continue
		}
		if err != nil {
			return "", errors.WithStack(err)
		}

		var (
			buf     []byte
			erd     error
			decoded bool
		)
		switch v.ValueFormat {
		case "rawurl-base64-json":
			buf, erd = base64.RawURLEncoding.DecodeString(ck.Value)
			if erd != nil {
				return "", errors.WithStack(errors.WithMessage(err, "cookie value expected as unpadded url base64"))
			}
			decoded = true
		case "rawstd-base64-json":
			fallthrough
		case "":
			buf, erd = base64.RawStdEncoding.DecodeString(ck.Value)
			if erd != nil {
				return "", errors.WithStack(errors.WithMessage(err, "cookie value expected as unpadded standard base64"))
			}
			decoded = true
		default:
			return "", errors.WithStack(ErrInvalidCookieFormat)
		}
		if decoded {
			var body map[string]json.RawMessage
			if erd := json.Unmarshal(buf, &body); erd != nil {
				return "", errors.WithStack(errors.WithMessage(erd, "cookie does not contain valid json value"))
			}
			var key string
			if v.TokenKey == "" {
				// default key in JSON cookie
				key = accessToken
			} else {
				key = v.TokenKey
			}
			str := body[key]
			if len(str) > 0 {
				if eru := json.Unmarshal(str, &token); eru != nil {
					return "", errors.WithStack(errors.WithMessage(eru, "token in cookie is not a json string"))
				}
			}
			if token != "" {
				break
			}
		}
	}
	return token, nil
}
