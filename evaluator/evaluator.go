package evaluator

import (
	"net/http"
)

type Evaluator interface {
	EvaluateAccessRequest(r *http.Request) (*Session, error)
}
