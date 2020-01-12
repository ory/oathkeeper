package rule

import (
	"hash/crc32"

	"github.com/dlclark/regexp2"
	"github.com/ory/ladon/compiler"
)

type regexpMatchingEngine struct {
	compiled *regexp2.Regexp
	checksum uint32
}

func (re *regexpMatchingEngine) compile(pattern string) error {
	if checksum := crc32.ChecksumIEEE([]byte(pattern)); checksum != re.checksum {
		compiled, err := compiler.CompileRegex(pattern, '<', '>')
		if err != nil {
			return err
		}
		re.compiled = compiled
		re.checksum = checksum
	}
	return nil
}

// Checksum of a saved pattern.
func (re *regexpMatchingEngine) Checksum() uint32 {
	return re.checksum
}

// IsMatching determines whether the input matches the pattern.
func (re *regexpMatchingEngine) IsMatching(pattern, matchAgainst string) (bool, error) {
	if err := re.compile(pattern); err != nil {
		return false, err
	}
	return re.compiled.MatchString(matchAgainst)
}

// ReplaceAllString replaces all matches in `input` with `replacement`.
func (re *regexpMatchingEngine) ReplaceAllString(pattern, input, replacement string) (string, error) {
	if err := re.compile(pattern); err != nil {
		return "", err
	}
	return re.compiled.Replace(input, replacement, -1, -1)
}
