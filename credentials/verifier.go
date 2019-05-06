package credentials

import (
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/ory/fosite"
)

type Verifier interface {
	Verify(
		ctx context.Context,
		token string,
		algorithms []string,
		issuers []string,
		audiences []string,
		ss fosite.ScopeStrategy,
	) (*jwt.Token, error)
}
