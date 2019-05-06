package credential

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/ory/fosite"
	"net/url"
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
	Scope []string
	KeyURLs       []url.URL
}
