// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package api

//lint:file-ignore U1000 Used to generate Swagger/OpenAPI definitions.

import "github.com/ory/oathkeeper/rule"

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

// swagger:parameters listRules
type swaggerListRulesParameters struct {
	// The maximum amount of rules returned.
	// in: query
	Limit int `json:"limit"`

	// The offset from where to start looking.
	// in: query
	Offset int `json:"offset"`
}

// swagger:parameters getRule
type swaggerGetRuleParameters struct {
	// in: path
	// required: true
	ID string `json:"id"`
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

	// Mutators is a list of mutation handlers that transform the HTTP request. A common use case is generating a new set
	// of credentials (e.g. JWT) which then will be forwarded to the upstream server.
	//
	// Mutations are performed iteratively from index 0 to n and should all succeed in order for the HTTP request to be forwarded.
	Mutators []swaggerRuleHandler `json:"mutators"`

	// Upstream is the location of the server where requests matching this rule should be forwarded to.
	Upstream *rule.Upstream `json:"upstream"`
}
