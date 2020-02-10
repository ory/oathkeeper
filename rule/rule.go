/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package rule

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/driver/configuration"
)

type Match struct {
	// An array of HTTP methods (e.g. GET, POST, PUT, DELETE, ...). When ORY Oathkeeper searches for rules
	// to decide what to do with an incoming request to the proxy server, it compares the HTTP method of the incoming
	// request with the HTTP methods of each rules. If a match is found, the rule is considered a partial match.
	// If the matchesUrl field is satisfied as well, the rule is considered a full match.
	Methods []string `json:"methods"`

	// This field represents the URL pattern this rule matches. When ORY Oathkeeper searches for rules
	// to decide what to do with an incoming request to the proxy server, it compares the full request URL
	// (e.g. https://mydomain.com/api/resource) without query parameters of the incoming
	// request with this field. If a match is found, the rule is considered a partial match.
	// If the matchesMethods field is satisfied as well, the rule is considered a full match.
	//
	// You can use regular expressions or glob patterns in this field to match more than one url.
	// The matching strategy is determined by configuration parameter MatchingStrategy.
	// Regular expressions and glob patterns are encapsulated in brackets < and >.
	// The following regexp example matches all paths of the domain `mydomain.com`: `https://mydomain.com/<.*>`.
	// The glob equivalent of the above regexp example is `https://mydomain.com/<*>`.
	URL string `json:"url"`
}

type Handler struct {
	// Handler identifies the implementation which will be used to handle this specific request. Please read the user
	// guide for a complete list of available handlers.
	Handler string `json:"handler"`

	// Config contains the configuration for the handler. Please read the user
	// guide for a complete list of each handler's available settings.
	Config json.RawMessage `json:"config"`
}

type ErrorHandler struct {
	// Handler identifies the implementation which will be used to handle this specific request. Please read the user
	// guide for a complete list of available handlers.
	Handler string `json:"handler"`

	// Config defines additional configuration for the response handler.
	Config json.RawMessage `json:"config"`
}

type OnErrorRequest struct {
	// ContentType defines the content type(s) that should match. Wildcards such as `application/*` are supported.
	ContentType []string `json:"content_type"`

	// Accept defines the accept header that should match. Wildcards such as `application/*` are supported.
	Accept []string `json:"accept"`
}

// Rule is a single rule that will get checked on every HTTP request.
type Rule struct {
	// ID is the unique id of the rule. It can be at most 190 characters long, but the layout of the ID is up to you.
	// You will need this ID later on to update or delete the rule.
	ID string `json:"id"`

	// Version represents the access rule version. Should match one of ORY Oathkeepers release versions. Supported since
	// v0.20.0-beta.1+oryOS.14.
	Version string `json:"version"`

	// Description is a human readable description of this rule.
	Description string `json:"description"`

	// Match defines the URL that this rule should match.
	Match *Match `json:"match"`

	// Authenticators is a list of authentication handlers that will try and authenticate the provided credentials.
	// Authenticators are checked iteratively from index 0 to n and if the first authenticator to return a positive
	// result will be the one used.
	//
	// If you want the rule to first check a specific authenticator  before "falling back" to others, have that authenticator
	// as the first item in the array.
	Authenticators []Handler `json:"authenticators"`

	// Authorizer is the authorization handler which will try to authorize the subject (authenticated using an Authenticator)
	// making the request.
	Authorizer Handler `json:"authorizer"`

	// Mutators is a list of mutation handlers that transform the HTTP request. A common use case is generating a new set
	// of credentials (e.g. JWT) which then will be forwarded to the upstream server.
	//
	// Mutations are performed iteratively from index 0 to n and should all succeed in order for the HTTP request to be forwarded.
	Mutators []Handler `json:"mutators"`

	// Errors is a list of error handlers. These will be invoked if any part of the system returns an error. You can
	// configure error matchers to listen on certain errors (e.g. unauthorized) and execute specific logic (e.g. redirect
	// to the login endpoint, return with an XML error, return a json error, ...).
	Errors []ErrorHandler `json:"errors"`

	// Upstream is the location of the server where requests matching this rule should be forwarded to.
	Upstream Upstream `json:"upstream"`

	matchingEngine MatchingEngine
}

type Upstream struct {
	// PreserveHost, if false (the default), tells ORY Oathkeeper to set the upstream request's Host header to the
	// hostname of the API's upstream's URL. Setting this flag to true instructs ORY Oathkeeper not to do so.
	PreserveHost bool `json:"preserve_host"`

	// StripPath if set, replaces the provided path prefix when forwarding the requested URL to the upstream URL.
	StripPath string `json:"strip_path"`

	// URL is the URL the request will be proxied to.
	URL string `json:"url"`
}

var _ json.Unmarshaler = new(Rule)

func (r *Rule) UnmarshalJSON(raw []byte) error {
	var rr struct {
		ID             string         `json:"id"`
		Version        string         `json:"version"`
		Description    string         `json:"description"`
		Match          *Match         `json:"match"`
		Authenticators []Handler      `json:"authenticators"`
		Authorizer     Handler        `json:"authorizer"`
		Mutators       []Handler      `json:"mutators"`
		Errors         []ErrorHandler `json:"errors"`
		Upstream       Upstream       `json:"upstream"`
		matchingEngine MatchingEngine
	}

	transformed, err := migrateRuleJSON(raw)
	if err != nil {
		return err
	}

	if err := errors.WithStack(json.Unmarshal(transformed, &rr)); err != nil {
		return err
	}

	*r = rr
	return nil
}

// GetID returns the rule's ID.
func (r *Rule) GetID() string {
	return r.ID
}

// IsMatching checks whether the provided url and method match the rule.
// An error will be returned if a regexp matching strategy is selected and regexp timeout occurs.
func (r *Rule) IsMatching(strategy configuration.MatchingStrategy, method string, u *url.URL) (bool, error) {
	if !stringInSlice(method, r.Match.Methods) {
		return false, nil
	}
	if err := ensureMatchingEngine(r, strategy); err != nil {
		return false, err
	}
	matchAgainst := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
	return r.matchingEngine.IsMatching(r.Match.URL, matchAgainst)
}

// ReplaceAllString searches the input string and replaces each match (with the rule's pattern)
// found with the replacement text.
func (r *Rule) ReplaceAllString(strategy configuration.MatchingStrategy, input, replacement string) (string, error) {
	if err := ensureMatchingEngine(r, strategy); err != nil {
		return "", err
	}

	return r.matchingEngine.ReplaceAllString(r.Match.URL, input, replacement)
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.EqualFold(a, b) {
			return true
		}
	}
	return false
}

func ensureMatchingEngine(rule *Rule, strategy configuration.MatchingStrategy) error {
	if rule.matchingEngine != nil {
		return nil
	}
	switch strategy {
	case configuration.Glob:
		rule.matchingEngine = new(globMatchingEngine)
		return nil
	case "", configuration.Regexp:
		rule.matchingEngine = new(regexpMatchingEngine)
		return nil
	}

	return errors.Wrap(ErrUnknownMatchingStrategy, string(strategy))
}

// ExtractRegexGroups returns the values matching the rule pattern
func (r *Rule) ExtractRegexGroups(strategy configuration.MatchingStrategy, u *url.URL) ([]string, error) {
	if err := ensureMatchingEngine(r, strategy); err != nil {
		return nil, err
	}

	if r.Match == nil {
		return []string{}, nil
	}

	matchAgainst := fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, u.Path)
	groups, err := r.matchingEngine.FindStringSubmatch(r.Match.URL, matchAgainst)
	if err != nil {
		return nil, err
	}

	return groups, nil
}
