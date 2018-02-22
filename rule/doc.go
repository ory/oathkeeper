// Package rule implements management capabilities for rules
//
// A rule is used to decide what to do with requests that are hitting the ORY Oathkeeper proxy server. A rule must
// define the HTTP methods and the URL under which it will apply. A URL may not have more than one rule. If a URL
// has no rule applied, the proxy server will return a 404 not found error.
//
// ORY Oathkeeper stores as many rules as required and iterates through them on every request. Rules are essential
// to the way ORY Oathkeeper works. To read more on rules, please refer to the developer guide: https://ory.gitbooks.io/oathkeeper/content/concepts.html#rules
// Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package rule

// A rule
// swagger:response rule
type swaggerRuleResponse struct {
	// in: body
	Body jsonRule
}

// A list of rules
// swagger:response rules
type swaggerRulesResponse struct {
	// in: body
	// type: array
	Body []jsonRule
}

// swagger:parameters getRule deleteRule
type swaggerGetRuleParameters struct {
	// in: path
	// required: true
	ID string `json:"id"`
}

// swagger:parameters updateRule
type swaggerUpdateRuleParameters struct {
	// in: path
	// required: true
	ID string `json:"id"`

	// in: body
	Body jsonRule
}

// swagger:parameters createRule
type swaggerCreateRuleParameters struct {
	// in: body
	Body jsonRule
}

// A rule
// swagger:model rule
type jsonRule struct {
	// The ID is the unique id of the rule. It can be at most 190 characters long, but the layout of the ID is up to you.
	// You will need this ID later on to update or delete the rule.
	ID string `json:"id" db:"id"`

	// A human readable description of this rule.
	Description string `json:"description"`

	// An array of HTTP methods (e.g. GET, POST, PUT, DELETE, ...). When ORY Oathkeeper searches for rules
	// to decide what to do with an incoming request to the proxy server, it compares the HTTP method of the incoming
	// request with the HTTP methods of each rules. If a match is found, the rule is considered a partial match.
	// If the matchesUrl field is satisfied as well, the rule is considered a full match.
	MatchesMethods []string `json:"matchesMethods"`

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
	MatchesURL string `json:"matchesUrl"`

	// An array of OAuth 2.0 scopes that are required when accessing an endpoint protected by this rule.
	// If the token used in the Authorization header did not request that specific scope, the request is denied.
	RequiredScopes []string `json:"requiredScopes"`

	// Defines which mode this rule should use. There are four valid modes:
	//
	// - bypass: If set, any authorization logic is completely disabled and the Authorization header is not changed at all.
	// 		This is useful if you have an endpoint that has it's own authorization logic, for example using basic authorization.
	// 		If set to true, this setting overrides `basicAuthorizationModeEnabled` and `allowAnonymousModeEnabled`.
	// - anonymous: If set, the protected endpoint is available to anonymous users. That means that the endpoint is accessible
	// 		without having a valid access token. This setting overrides `basicAuthorizationModeEnabled`.
	// - token: If set, disables checks against ORY Hydra's Warden API and uses basic authorization. This means that
	// 		the access token is validated (e.g. checking if it is expired, check if it claimed the necessary scopes)
	// 		but does not use the `requiredAction` and `requiredResource` fields for advanced access control.
	// - policy: If set, uses ORY Hydra's Warden API for access control using access control policies.
	Mode string `json:"mode"`

	// This field will be used to decide advanced authorization requests where access control policies are used. A
	// action is typically something a user wants to do (e.g. write, read, delete).
	// This field supports expansion as described in the developer guide: https://ory.gitbooks.io/oathkeeper/content/concepts.html#rules
	RequiredAction string `json:"requiredAction"`

	// This field will be used to decide advanced authorization requests where access control policies are used. A
	// resource is typically something a user wants to access (e.g. printer, article, virtual machine).
	// This field supports expansion as described in the developer guide: https://ory.gitbooks.io/oathkeeper/content/concepts.html#rules
	RequiredResource string `json:"requiredResource"`
}
