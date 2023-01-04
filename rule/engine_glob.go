// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"bytes"
	"hash/crc64"

	"github.com/gobwas/glob"
)

type globMatchingEngine struct {
	compiled glob.Glob
	checksum uint64
	table    *crc64.Table
}

// Checksum of a saved pattern.
func (ge *globMatchingEngine) Checksum() uint64 {
	return ge.checksum
}

// IsMatching determines whether the input matches the pattern.
func (ge *globMatchingEngine) IsMatching(pattern, matchAgainst string) (bool, error) {
	if err := ge.compile(pattern); err != nil {
		return false, err
	}
	return ge.compiled.Match(matchAgainst), nil
}

// ReplaceAllString is noop for now and always returns an error.
func (ge *globMatchingEngine) ReplaceAllString(_, _, _ string) (string, error) {
	return "", ErrMethodNotImplemented
}

// FindStringSubmatch is noop for now and always returns an empty array
func (ge *globMatchingEngine) FindStringSubmatch(pattern, matchAgainst string) ([]string, error) {
	return []string{}, nil
}

func (ge *globMatchingEngine) compile(pattern string) error {
	if ge.table == nil {
		ge.table = crc64.MakeTable(polynomial)
	}
	if checksum := crc64.Checksum([]byte(pattern), ge.table); checksum != ge.checksum {
		compiled, err := compileGlob(pattern, '<', '>')
		if err != nil {
			return err
		}
		ge.checksum = checksum
		ge.compiled = compiled
	}
	return nil
}

// delimiterIndices returns the first level delimiter indices from a string.
// It returns an error in case of unbalanced delimiters.
func delimiterIndices(s string, delimiterStart, delimiterEnd rune) ([]int, error) {
	var level, idx int
	idxs := make([]int, 0)
	for ind := 0; ind < len(s); ind++ {
		switch s[ind] {
		case byte(delimiterStart):
			if level++; level == 1 {
				idx = ind
			}
		case byte(delimiterEnd):
			if level--; level == 0 {
				idxs = append(idxs, idx, ind+1)
			} else if level < 0 {
				return nil, ErrUnbalancedPattern
			}
		}
	}

	if level != 0 {
		return nil, ErrUnbalancedPattern
	}
	return idxs, nil
}

func compileGlob(pattern string, delimiterStart, delimiterEnd rune) (glob.Glob, error) {
	// Check if it is well-formed.
	idxs, errBraces := delimiterIndices(pattern, delimiterStart, delimiterEnd)
	if errBraces != nil {
		return nil, errBraces
	}
	buffer := bytes.NewBufferString("")

	var end int
	for ind := 0; ind < len(idxs); ind += 2 {
		// Set all values we are interested in.
		raw := pattern[end:idxs[ind]]
		end = idxs[ind+1]
		patt := pattern[idxs[ind]+1 : end-1]
		buffer.WriteString(glob.QuoteMeta(raw))
		buffer.WriteString(patt)
	}

	// Add the remaining.
	raw := pattern[end:]
	buffer.WriteString(glob.QuoteMeta(raw))

	// Compile full regexp.
	return glob.Compile(buffer.String(), '.', '/')
}
