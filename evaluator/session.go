package evaluator

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Session struct {
	User      string `json:"user"`
	Anonymous bool   `json:"anonymous"`
	Disabled  bool   `json:"disabled"`
	ClientID  string `json:"clientId"`
}

func (s *Session) ToClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"nbf":  time.Now().Unix(),
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
		"sub":  s.User,
		"anon": s.Anonymous,
		"cid":  s.ClientID,
	}
}
