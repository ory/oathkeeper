// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package credentials

import (
	"context"
	"net/url"

	"github.com/golang-jwt/jwt/v5"
)

type Signer interface {
	Sign(ctx context.Context, location *url.URL, claims jwt.Claims) (string, error)
}

type SignerRegistry interface {
	CredentialsSigner() Signer
}
