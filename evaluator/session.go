// Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package evaluator

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pborman/uuid"
)

type Session struct {
	User      string      `json:"user"`
	Anonymous bool        `json:"anonymous"`
	Disabled  bool        `json:"disabled"`
	ClientID  string      `json:"clientId"`
	Issuer    string      `json:"issuer"`
	Extra     interface{} `json:"extra"`
}

func (s *Session) ToClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"nbf":  time.Now().Unix(),
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
		"sub":  s.User,
		"iss":  s.Issuer,
		"anon": s.Anonymous,
		"aud":  s.ClientID,
		"jti":  uuid.New(),
		"ext":  s.Extra,
	}
}
