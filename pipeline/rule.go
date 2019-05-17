package pipeline

import "regexp"

type Rule interface {
	GetID() string
	CompileURL() (*regexp.Regexp, error)
}
