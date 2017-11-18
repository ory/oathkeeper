package evaluator

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pborman/uuid"
)

type Session struct {
	User      string `json:"user"`
	Anonymous bool   `json:"anonymous"`
	Disabled  bool   `json:"disabled"`
	ClientID  string `json:"clientId"`
	Issuer    string `json:"issuer"`
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
	}
}
