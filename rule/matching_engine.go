package rule

import (
	"github.com/pkg/errors"
)

// common errors for MatchingEngine.
var (
	ErrUnbalancedPattern       = errors.New("unbalanced pattern")
	ErrMethodNotImplemented    = errors.New("the method is not implemented")
	ErrUnknownMatchingStrategy = errors.New("unknown matching strategy")
)

// MatchingEngine describes an interface of matching engine such as regexp or glob.
type MatchingEngine interface {
	IsMatching(pattern, matchAgainst string) (bool, error)
	ReplaceAllString(pattern, input, replacement string) (string, error)
	Checksum() uint32
}
