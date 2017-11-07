// package rule encapsulates rule management logic as well as rule matching logic.
//
//
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
	// ID the a unique id of a rule.
	ID string `json:"id" db:"id"`

	// MatchesMethods is a list of HTTP methods that this rule matches.
	MatchesMethods []string `json:"matchesMethods"`

	// MatchesPathCompiled is a regular expression of paths this rule matches.
	MatchesPath string `json:"matchesPath"`

	// RequiredScopes is a list of scopes that are required by this rule.
	RequiredScopes []string `json:"requiredScopes"`

	// RequiredScopes is the action this rule requires.
	RequiredAction string `json:"requiredAction"`

	// RequiredScopes is the resource this rule requires.
	RequiredResource string `json:"requiredResource"`

	// AllowAnonymous sets if the endpoint is public, thus not needing any authorization at all.
	AllowAnonymous bool `json:"allowAnonymous"`

	// Description describes the rule.
	Description string `json:"description"`

	// BypassAuthorization if set true disables firewall capabilities.
	BypassAuthorization bool `json:"bypassAuthorization"`

	// BypassAccessControlPolicies if set true disables checking access control policies.
	BypassAccessControlPolicies bool `json:"bypassAccessControlPolicies"`
}
