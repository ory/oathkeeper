// Copyright 2021 Ory GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package credentials

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"net/url"
	"reflect"

	"github.com/form3tech-oss/jwt-go"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
	"gopkg.in/square/go-jose.v2"
)

var _ Signer = new(DefaultSigner)

type DefaultSigner struct {
	r FetcherRegistry
}

func NewSignerDefault(r FetcherRegistry) *DefaultSigner {
	return &DefaultSigner{r: r}
}

func (s *DefaultSigner) Sign(ctx context.Context, location *url.URL, claims jwt.Claims) (string, error) {
	key, id, err := s.key(ctx, location)
	if err != nil {
		return "", err
	}

	method := jwt.GetSigningMethod(key.Algorithm)
	if method == nil {
		return "", errors.Errorf(`credentials: signing key "%s" declares unsupported algorithm "%s"`, key.KeyID, key.Algorithm)
	}

	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = id

	signed, err := token.SignedString(key.Key)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return signed, nil
}

func (s *DefaultSigner) key(ctx context.Context, location *url.URL) (*jose.JSONWebKey, string, error) {
	keys, err := s.r.CredentialsFetcher().ResolveSets(ctx, []url.URL{*location})
	if err != nil {
		return nil, "", err
	}

	if len(keys) != 1 {
		return nil, "", errors.Errorf("credentials: expected exactly one JSON Web Key Set to be returned but got: %d", len(keys))
	}

	var pk jose.JSONWebKey
	var kid string
	for _, key := range keys[0].Keys {
		switch key.Key.(type) {
		case ed25519.PrivateKey:
			pk = key
		case ed25519.PublicKey:
			kid = key.KeyID

		case *ecdsa.PrivateKey:
			pk = key
		case *ecdsa.PublicKey:
			kid = key.KeyID

		case *rsa.PrivateKey:
			pk = key
		case *rsa.PublicKey:
			kid = key.KeyID

		case []byte:
			pk = key
			kid = key.KeyID

		default:
			return nil, "", errors.Errorf("credentials: unknown key type '%s'", reflect.TypeOf(key))
		}

		if pk.Key != nil && kid != "" {
			break
		}
	}

	if pk.KeyID == "" {
		return nil, "", errors.Errorf("credentials: no suitable key could be found")
	}

	if kid == "" {
		kid = pk.KeyID
	}

	return &pk, kid, nil
}
