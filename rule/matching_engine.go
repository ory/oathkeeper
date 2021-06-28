// Copyright 2021 Ory GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
