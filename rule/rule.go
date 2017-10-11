package rule

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// Rule is a single rule that will get checked on every HTTP request.
type Rule struct {
	// ID the a unique id of a rule.
	ID string

	// MatchesMethods is a list of HTTP methods that this rule matches.
	MatchesMethods []string

	// MatchesPath is a regular expression of paths this rule matches.
	MatchesPath *regexp.Regexp

	// RequiredScopes is a list of scopes that are required by this rule.
	RequiredScopes []string

	// RequiredScopes is the action this rule requires.
	RequiredAction string

	// RequiredScopes is the resource this rule requires.
	RequiredResource string

	// AllowAnonymous if set true allows anonymous access to the specified paths and methods.
	AllowAnonymous bool

	// BypassAuthorization if set true disables firewall capabilities.
	BypassAuthorization bool

	// BypassAccessControlPolicies if set true disables checking access control policies.
	BypassAccessControlPolicies bool

	// Description describes the rule.
	Description string
}

func (r *Rule) MatchesURL(method string, u *url.URL) error {
	if !stringInSlice(method, r.MatchesMethods) {
		return errors.Errorf("Method %s does not match any of %v", method, r.MatchesMethods)
	}

	if !r.MatchesPath.MatchString(u.Path) {
		return errors.Errorf("Path %s does not match %s", u.Path, r.MatchesPath.String())
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
