package pipeline

import (
	"github.com/ory/oathkeeper/driver/configuration"
)

type Rule interface {
	GetID() string
	// ReplaceAllString searches the input string and replaces each match (with the rule's pattern)
	// found with the replacement text.
	ReplaceAllString(strategy configuration.MatchingStrategy, input, replacement string) (string, error)
}
