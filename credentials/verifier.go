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

package credentials

import (
	"context"
	"net/url"

	"github.com/form3tech-oss/jwt-go"

	"github.com/ory/fosite"
)

type Verifier interface {
	Verify(
		ctx context.Context,
		token string,
		r *ValidationContext,
	) (*jwt.Token, error)
}

type VerifierRegistry interface {
	CredentialsVerifier() Verifier
}

type ValidationContext struct {
	Algorithms    []string
	Issuers       []string
	Audiences     []string
	ScopeStrategy fosite.ScopeStrategy
	Scope         []string
	KeyURLs       []url.URL
}
