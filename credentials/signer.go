package credentials

import (
	"github.com/dgrijalva/jwt-go"
)

type Signer interface {
	Sign(claims jwt.Claims) (string, error)
}
