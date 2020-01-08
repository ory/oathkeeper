package pipeline

import (
	"github.com/dlclark/regexp2"
)

type Rule interface {
	GetID() string
	CompileURL() (*regexp2.Regexp, error)
}
