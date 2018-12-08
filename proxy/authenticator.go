package proxy

import (
	"encoding/json"
	"net/http"

	"github.com/go-errors/errors"

	"github.com/ory/oathkeeper/rule"
)

var ErrAuthenticatorNotResponsible = errors.New("Authenticator not responsible")
var ErrAuthenticatorBypassed = errors.New("Authenticator is disabled")

type Authenticator interface {
	Authenticate(r *http.Request, config json.RawMessage, rl *rule.Rule) (*AuthenticationSession, error)
	GetID() string
}

type AuthenticationSession struct {
	Subject string
	Extra   map[string]interface{}
}

//
//func (s *Default*AuthenticationSession) ToClaims() map[string]interface{} {
//	return map[string]interface{}{
//		"nbf":  time.Now().Unix(),
//		"iat":  time.Now().Unix(),
//		"exp":  time.Now().Add(time.Hour).Unix(),
//		"sub":  s.Subject,
//		"iss":  s.Issuer,
//		"anon": s.Anonymous,
//		"aud":  s.ClientID,
//		"jti":  uuid.New(),
//		"ext":  s.Extra,
//	}
//}
