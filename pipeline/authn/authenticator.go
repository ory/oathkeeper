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

package authn

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/ory/herodot"

	"github.com/mitchellh/copystructure"

	"github.com/ory/oathkeeper/pipeline"
)

var ErrAuthenticatorNotResponsible = errors.New("Authenticator not responsible")
var ErrAuthenticatorNotEnabled = herodot.DefaultError{
	ErrorField:  "authenticator matching this route is misconfigured or disabled",
	CodeField:   http.StatusInternalServerError,
	StatusField: http.StatusText(http.StatusInternalServerError),
}

type Authenticator interface {
	Authenticate(r *http.Request, session *AuthenticationSession, config json.RawMessage, rule pipeline.Rule) error
	GetID() string
	Validate(config json.RawMessage) error
}

func NewErrAuthenticatorNotEnabled(a Authenticator) *herodot.DefaultError {
	return ErrAuthenticatorNotEnabled.WithTrace(errors.New("")).WithReasonf(`Authenticator "%s" is disabled per configuration.`, a.GetID())
}

func NewErrAuthenticatorMisconfigured(a Authenticator, err error) *herodot.DefaultError {
	return ErrAuthenticatorNotEnabled.WithTrace(err).WithReasonf(
		`Configuration for authenticator "%s" could not be validated: %s`,
		a.GetID(),
		err,
	)
}

type AuthenticationSession struct {
	Subject      string                 `json:"subject"`
	Extra        map[string]interface{} `json:"extra"`
	Header       http.Header            `json:"header"`
	MatchContext MatchContext           `json:"match_context"`
}

type MatchContext struct {
	RegexpCaptureGroups []string    `json:"regexp_capture_groups"`
	URL                 *url.URL    `json:"url"`
	Method              string      `json:"method"`
	Header              http.Header `json:"header"`
}

func (a *AuthenticationSession) SetHeader(key, val string) {
	if a.Header == nil {
		a.Header = map[string][]string{}
	}
	a.Header.Set(key, val)
}

func (a *AuthenticationSession) Copy() *AuthenticationSession {
	return copystructure.Must(copystructure.Copy(a)).(*AuthenticationSession)
}
