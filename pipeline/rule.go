package pipeline

type Rule interface {
	GetID() string
	// Replace searches the input string and replaces each match (with the rule's pattern)
	// found with the replacement text.
	Replace(input, replacement string) (string, error)
}
