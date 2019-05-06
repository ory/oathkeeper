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
	"strings"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-errors/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/rsakey"
)

func TestCredentialsIssuerIDToken(t *testing.T) {
	var keys = map[string]rsakey.Manager{
		"hs256": rsakey.NewLocalHS256Manager([]byte("foobarbaz")),
		"rs256": &rsakey.LocalRS256Manager{KeyStrength: 512},
	}

	for m, k := range keys {
		t.Run(fmt.Sprintf("algo=%s", m), func(t *testing.T) {
			b := NewCredentialsIssuerIDToken(k, logrus.New(), time.Hour, "some-issuer")

			assert.NotNil(t, b)
			assert.NotEmpty(t, b.GetID())

			r := &http.Request{Header: http.Header{}}
			s := &AuthenticationSession{Subject: "foo"}

			header, err := b.Transform(r, s, json.RawMessage([]byte(`{ "aud": ["foo", "bar"] }`)), nil)
			require.NoError(t, err)
			r.Header = header

			generated := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", 1)
			token, err := jwt.ParseWithClaims(generated, new(claims), func(token *jwt.Token) (interface{}, error) {
				_, rsa := token.Method.(*jwt.SigningMethodRSA)
				_, hmac := token.Method.(*jwt.SigningMethodHMAC)
				if !rsa && !hmac {
					return nil, errors.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				return k.PublicKey()
			})
			require.NoError(t, err)

			claims := token.Claims.(*claims)
			assert.Equal(t, "foo", claims.Subject)
			assert.Equal(t, []string{"foo", "bar"}, claims.Audience)
		})
	}
}
