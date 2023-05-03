// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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

type AuthenticatorForwardConfig interface {
	GetCheckSessionURL() string
	GetPreserveQuery() bool
	GetPreservePath() bool
	GetPreserveHost() bool
	GetForwardHTTPHeaders() []string
	GetSetHeaders() map[string]string
	GetForceMethod() string
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
