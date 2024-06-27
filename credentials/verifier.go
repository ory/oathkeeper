// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package credentials

import (
	"context"
	"net/url"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ory/fosite"
)

type Verifier interface {
	Verify(
		ctx context.Context,
		token string,
		r *ValidationContext,
	) (*jwt.Token, error)
}

type VerifierRegistry interface {
	CredentialsVerifier() Verifier
}

type ValidationContext struct {
	Algorithms    []string
	Issuers       []string
	Audiences     []string
	ScopeStrategy fosite.ScopeStrategy
	Scope         []string
	KeyURLs       []url.URL
}
