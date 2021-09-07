package credentials

import (
	"context"
	"net/url"

	"github.com/golang-jwt/jwt/v4"

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
