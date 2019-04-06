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

// Package rule implements management capabilities for rules
//
// A rule is used to decide what to do with requests that are hitting the ORY Oathkeeper proxy server. A rule must
// define the HTTP methods and the URL under which it will apply. A URL may not have more than one rule. If a URL
// has no rule applied, the proxy server will return a 404 not found error.
//
// ORY Oathkeeper stores as many rules as required and iterates through them on every request. Rules are essential
// to the way ORY Oathkeeper works.
package rule

// A rule
// swagger:response rule
type swaggerRuleResponse struct {
	// in: body
	Body swaggerRule
}

// A list of rules
// swagger:response rules
type swaggerRulesResponse struct {
	// in: body
	// type: array
	Body []swaggerRule
}

// swagger:parameters getRule deleteRule
type swaggerGetRuleParameters struct {
	// in: path
	// required: true
	ID string `json:"id"`
}

// swagger:parameters listRules
type swaggerListRulesParameters struct {
	// The maximum amount of rules returned.
	// in: query
	Limit int `json:"limit"`

	// The offset from where to start looking.
	// in: query
	Offset int `json:"offset"`
}

// swagger:parameters updateRule
type swaggerUpdateRuleParameters struct {
	// in: path
	// required: true
	ID string `json:"id"`

	// in: body
	Body swaggerRule
}

// swagger:parameters createRule
type swaggerCreateRuleParameters struct {
	// in: body
	Body swaggerRule
}

// swagger:model ruleMatch
type swaggerRuleMatch struct {
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
	URL string `json:"url"`
}

// swagger:model ruleHandler
type swaggerRuleHandler struct {
	// Handler identifies the implementation which will be used to handle this specific request. Please read the user
	// guide for a complete list of available handlers.
	Handler string `json:"handler"`

	// Config contains the configuration for the handler. Please read the user
	// guide for a complete list of each handler's available settings.
	Config interface{} `json:"config"`
}

// swaggerRule is a single rule that will get checked on every HTTP request.
// swagger:model rule
type swaggerRule struct {
	// ID is the unique id of the rule. It can be at most 190 characters long, but the layout of the ID is up to you.
	// You will need this ID later on to update or delete the rule.
	ID string `json:"id" db:"surrogate_id"`

	// Description is a human readable description of this rule.
	Description string `json:"description"`

	// Match defines the URL that this rule should match.
	Match swaggerRuleMatch `json:"match"`

	// Authenticators is a list of authentication handlers that will try and authenticate the provided credentials.
	// Authenticators are checked iteratively from index 0 to n and if the first authenticator to return a positive
	// result will be the one used.
	//
	// If you want the rule to first check a specific authenticator  before "falling back" to others, have that authenticator
	// as the first item in the array.
	Authenticators []swaggerRuleHandler `json:"authenticators"`

	// Authorizer is the authorization handler which will try to authorize the subject (authenticated using an Authenticator)
	// making the request.
	Authorizer swaggerRuleHandler `json:"authorizer"`

	// CredentialsIssuer is the handler which will issue the credentials which will be used when ORY Oathkeeper
	// forwards a granted request to the upstream server.
	CredentialsIssuer swaggerRuleHandler `json:"credentials_issuer"`

	// Upstream is the location of the server where requests matching this rule should be forwarded to.
	Upstream *Upstream `json:"upstream"`
}
