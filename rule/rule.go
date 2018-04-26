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
	"net/url"
	"strings"

	"encoding/json"
	"regexp"

	"github.com/ory/ladon/compiler"
	"github.com/pkg/errors"
)

type RuleMatch struct {
	PreserveHost bool     `json:"preserveHost"`
	StripPath    string   `json:"stripPath"`


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
	// You can use regular expressions in this field to match more than one url. Regular expressions are encapsulated in
	// brackets < and >. The following example matches all paths of the domain `mydomain.com`: `https://mydomain.com/<.*>`.
	//
	// For more information refer to: https://ory.gitbooks.io/oathkeeper/content/concepts.html#rules
	URL string `json:"url"`

	Upstream *Upstream `json:"upstream"`
}

type RuleHandler struct {
	Handler string `json:"handler"`
	Config json.RawMessage `json:"config"`
}

// Rule is a single rule that will get checked on every HTTP request.
// swagger:model rule
type Rule struct {
	// The ID is the unique id of the rule. It can be at most 190 characters long, but the layout of the ID is up to you.
	// You will need this ID later on to update or delete the rule.
	ID string `json:"id"`

	// A human readable description of this rule.
	Description string `json:"description"`

	Matches []RuleMatch

	Authenticators map[string]RuleHandler
	Authorizer RuleHandler
	Session RuleHandler

	// This field will be used to decide advanced authorization requests where access control policies are used. A
	// action is typically something a user wants to do (e.g. write, read, delete).
	// This field supports expansion as described in the developer guide: https://ory.gitbooks.io/oathkeeper/content/concepts.html#rules
	RequiredAction string `json:"requiredAction"`

	// This field will be used to decide advanced authorization requests where access control policies are used. A
	// resource is typically something a user wants to access (e.g. printer, article, virtual machine).
	// This field supports expansion as described in the developer guide: https://ory.gitbooks.io/oathkeeper/content/concepts.html#rules
	RequiredResource string `json:"requiredResource"`
}

type jsonRule struct {
	ID               string    `json:"id"`
	Description      string    `json:"description"`
	MatchesMethods   []string  `json:"matchesMethods"`
	MatchesURL       string    `json:"matchesUrl"`
	RequiredScopes   []string  `json:"requiredScopes"`
	Mode             string    `json:"mode"`
	RequiredAction   string    `json:"requiredAction"`
	RequiredResource string    `json:"requiredResource"`
	Upstream         *Upstream `json:"upstream"`
}

type Upstream struct {
	URL          string   `json:"url"`
	URLParsed    *url.URL `json:"-"`
	PreserveHost bool     `json:"preserveHost"`
	StripPath    string   `json:"stripPath"`
}

func (r *Rule) UnmarshalJSON(data []byte) (err error) {
	f := &jsonRule{
		Upstream: new(Upstream),
	}
	if err = json.Unmarshal(data, f); err != nil {
		return errors.WithStack(err)
	}

	if f.Upstream == nil {
		f.Upstream = new(Upstream)
	}

	r.ID = f.ID
	r.Description = f.Description
	r.MatchesMethods = f.MatchesMethods
	r.MatchesURL = f.MatchesURL
	r.RequiredScopes = f.RequiredScopes
	r.Mode = f.Mode
	r.RequiredAction = f.RequiredAction
	r.RequiredResource = f.RequiredResource
	r.Upstream = &Upstream{
		URL:          f.Upstream.URL,
		PreserveHost: f.Upstream.PreserveHost,
		StripPath:    f.Upstream.StripPath,
	}

	if r.RequiredScopes == nil {
		r.RequiredScopes = []string{}
	}

	if r.MatchesMethods == nil {
		r.MatchesMethods = []string{}
	}

	if r.MatchesURLCompiled, err = compiler.CompileRegex(r.MatchesURL, '<', '>'); err != nil {
		return errors.WithStack(err)
	}

	if r.Upstream.URLParsed, err = url.Parse(r.Upstream.URL); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Rule) IsMatching(method string, u *url.URL) error {
	if !stringInSlice(method, r.MatchesMethods) {
		return errors.Errorf("Method %s does not match any of %v", method, r.MatchesMethods)
	}

	if !r.MatchesURLCompiled.MatchString(u.String()) {
		return errors.Errorf("Path %s does not match %s", u.String(), r.MatchesURL)
	}

	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.EqualFold(a, b) {
			return true
		}
	}
	return false
}
