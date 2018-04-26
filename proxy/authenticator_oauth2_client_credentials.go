package proxy

type AuthenticatorOAuth2ClientCredentialsConfiguration struct {
	// An array of OAuth 2.0 scopes that are required when accessing an endpoint protected by this rule.
	// If the token used in the Authorization header did not request that specific scope, the request is denied.
	RequiredScopes []string `json:"requiredScopes"`
}
