// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"hash/crc64"

	"github.com/pkg/errors"
)

// polynomial for crc64 table which is used for checking crc64 checksum
const polynomial = crc64.ECMA

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
	FindStringSubmatch(pattern, matchAgainst string) ([]string, error)
	Checksum() uint64
}
