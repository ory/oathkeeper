package proxy

import "net/http"

type Authorizer interface {
	Authorize(r *http.Request, session *AuthenticationSession) error
}
