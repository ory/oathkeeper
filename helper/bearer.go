// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
			if r.URL == nil {
				return ""
			}
			return r.URL.Query().Get(*tokenLocation.QueryParameter)
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
