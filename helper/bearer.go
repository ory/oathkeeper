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
	"net/http"
	"strings"
)

const (
	defaultAuthorizationHeader = "Authorization"
)

type BearerTokenLocation struct {
	Header         *string `json:"header"`
	QueryParameter *string `json:"query_parameter"`
	Cookie         *string `json:"cookie"`
}

func BearerTokenFromRequest(r *http.Request, tokenLocation *BearerTokenLocation) string {
	if tokenLocation != nil {
		if tokenLocation.Header != nil {
			if *tokenLocation.Header == defaultAuthorizationHeader {
				return DefaultBearerTokenFromRequest(r)
			}
			return r.Header.Get(*tokenLocation.Header)
		} else if tokenLocation.QueryParameter != nil {
			return r.FormValue(*tokenLocation.QueryParameter)
		} else if tokenLocation.Cookie != nil {
			cookie, err := r.Cookie(*tokenLocation.Cookie)
			if err != nil {
				return ""
			}
			return cookie.Value
		}
	}

	return DefaultBearerTokenFromRequest(r)
}

func DefaultBearerTokenFromRequest(r *http.Request) string {
	token := r.Header.Get(defaultAuthorizationHeader)
	split := strings.SplitN(token, " ", 2)
	if len(split) != 2 || !strings.EqualFold(strings.ToLower(split[0]), "bearer") {
		return ""
	}
	return split[1]
}
