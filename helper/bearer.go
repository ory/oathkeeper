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
	AuthScheme     *string `json:"auth_scheme"`
	QueryParameter *string `json:"query_parameter"`
	Cookie         *string `json:"cookie"`
}

func BearerTokenFromRequest(r *http.Request, tokenLocation *BearerTokenLocation) string {
	if tokenLocation != nil {
		if tokenLocation.Header != nil {
			if *tokenLocation.Header == defaultAuthorizationHeader {
				authScheme := "Bearer"
				if tokenLocation.AuthScheme != nil {
					authScheme = *tokenLocation.AuthScheme
				}
				return DefaultBearerTokenFromRequest(r, authScheme)
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

	return DefaultBearerTokenFromRequest(r, "Bearer")
}

func DefaultBearerTokenFromRequest(r *http.Request, authScheme string) string {
	token := r.Header.Get(defaultAuthorizationHeader)
	if authScheme == "" {
		return token
	}

	split := strings.SplitN(token, " ", 2)
	if len(split) != 2 || !strings.EqualFold(strings.ToLower(split[0]), strings.ToLower(authScheme)) {
		return ""
	}
	return split[1]
}
