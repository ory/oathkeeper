// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"errors"
	"hash/crc64"

	"github.com/dlclark/regexp2"

	"github.com/ory/ladon/compiler"
)

type regexpMatchingEngine struct {
	compiled *regexp2.Regexp
	checksum uint64
	table    *crc64.Table
}

func (re *regexpMatchingEngine) compile(pattern string) error {
	if re.table == nil {
		re.table = crc64.MakeTable(polynomial)
	}
	if checksum := crc64.Checksum([]byte(pattern), re.table); checksum != re.checksum {
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
func (re *regexpMatchingEngine) Checksum() uint64 {
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

// FindStringSubmatch returns all captures in matchAgainst following the pattern
func (re *regexpMatchingEngine) FindStringSubmatch(pattern, matchAgainst string) ([]string, error) {
	if err := re.compile(pattern); err != nil {
		return nil, err
	}

	m, _ := re.compiled.FindStringMatch(matchAgainst)
	if m == nil {
		return nil, errors.New("not match")
	}

	result := []string{}
	for _, group := range m.Groups()[1:] {
		result = append(result, group.String())
	}

	return result, nil
}
