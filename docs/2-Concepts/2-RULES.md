# Rules

<!-- toc -->

ORY Oathkeeper has a configurable set of rules. Rules are applied to all incoming requests and based on the rule definition,
an action is taken. There are four types of rules:

1. **Bypass**: Forwards the original request to the backend url without any modification to its headers.
2. **Anonymous**: Tries to extract user information from the given access credentials. If that fails, or no access
    credentials have been provided, the request is forwarded and the user is marked as "anonymous".
3. **Authenticated**: Requires valid access credentials and optionally checks for a set of OAuth 2.0 Scopes. If
    the supplied access credentials are invalid (expired, malformed, revoked) or do not fulfill the requested scopes,
    access is denied.
4. **Policy Based Access Control**: Requires valid access credentials as defined in 3. and additionally validates if the user
    is authorized to make the request, based on access control policies.

In general, a rule has the following fields:

* `id`: The unique id of the rule. It can be at most 190 characters long, but the layout of the ID is up to you. You will need this ID later on to update or delete the rule.
* `description`: A human readable description of this rule.
* `matchesMethods`: An array of HTTP methods (e.g. GET, POST, PUT, DELETE, ...). When ORY Oathkeeper searches for rules
    to decide what to do with an incoming request to the proxy server, it compares the HTTP method of the incoming
	request with the HTTP methods of each rules. If a match is found, the rule is considered a partial match.
	If the matchesUrl field is satisfied as well, the rule is considered a full match.
* `matchesUrl`: This field represents the URL pattern this rule matches. When ORY Oathkeeper searches for rules
	to decide what to do with an incoming request to the proxy server, it compares the full request URL
	(e.g. https://mydomain.com/api/resource) without query parameters of the incoming
	request with this field. If a match is found, the rule is considered a partial match.
	If the matchesMethods field is satisfied as well, the rule is considered a full match.

	You can use regular expressions in this field to match more than one url. Regular expressions are encapsulated in
	brackets < and >. The following example matches all paths of the domain `mydomain.com`: `https://mydomain.com/<.*>`.
* `requiredScopes`:  An array of OAuth 2.0 scopes that are required when accessing an endpoint protected by this rule.
	If the token used in the Authorization header did not request that specific scope, the request is denied.
* `modes`: Defines which mode this rule should use. There are four valid modes:
  * `bypass`: If set, any authorization logic is completely disabled and the Authorization header is not changed at all.
		This is useful if you have an endpoint that has it's own authorization logic, for example using basic authorization.
 		If set to true, this setting overrides `basicAuthorizationModeEnabled` and `allowAnonymousModeEnabled`.
  * `anonymous`: If set, the protected endpoint is available to anonymous users. That means that the endpoint is accessible
 		without having a valid access token. This setting overrides `basicAuthorizationModeEnabled`.
  * `token`: If set, disables checks against ORY Hydra's Warden API and uses basic authorization. This means that
 		the access token is validated (e.g. checking if it is expired, check if it claimed the necessary scopes)
 		but does not use the `requiredAction` and `requiredResource` fields for advanced access control.
  * `policy`: If set, uses ORY Hydra's Warden API for access control using access control policies.
        Mode string `json:"mode"`
* `requiredAction`: This field will be used to decide advanced authorization requests where access control policies are used. A
	action is typically something a user wants to do (e.g. write, read, delete).
	This field supports expansion as described in the developer guide: https://ory.gitbooks.io/oathkeeper/content/concepts.html#rules
* `requiredResource`: This field will be used to decide advanced authorization requests where access control policies are used. A
	resource is typically something a user wants to access (e.g. printer, article, virtual machine).
	This field supports expansion as described in the developer guide: https://ory.gitbooks.io/oathkeeper/content/concepts.html#rules

## Regular Expressions

Rules are matched using the `matchesUrl` parameter. That string is checked for case-sensitive equality:

* `http://localhost/users`
  * Matches: `http://localhost/users`
  * Does not match: `http://localhost/uSeRs`
  * Does not match: `http://localhost/users/1234`
* `http://localhost/users/1234`
  * Matches: `http://localhost/users/1234`
  * Does not match: `http://localhost/users`

In some scenarios that is not enough. For those cases, ORY Oathkeeper supports adding regular expressions to `matchesUrl`.
Regular expressions must be encapsulated in brackets `<` `>`, for example:

* `http://localhost/users/<[0-9]+>`
  * Matches: `http://localhost/users/1234`
  * Matches: `http://localhost/users/1235`
  * Does not match: `http://localhost/users/`
  * Does not match: `http://localhost/users/abc`
* `http://localhost/<.*>`
  * Matches: `http://localhost/users/1234`
  * Matches: `http://localhost/users`
  * Matches: `http://localhost/`
  * Does not match: `http://domain.com/users`

### Substitution

Sometimes it is required to use a part from the URL in the resource or action identifier. Let's assume that user
`subjects:mydomain.com:alexa` needs access to all articles beginning with the letter `0`, as defined by the following
policy:

```json
{
  "id": "some-policy-id",
  "subjects": [ "subjects:mydomain.com:alexa" ],
  "effect": "allow",
  "resources": [
    "resources:mydomain.com:articles:<0.+>"
  ],
  "actions": [ "get" ]
}
```

Now we need to use the article ID from the URL in the `requiredResource`. This can be achieved using [regular
expression group substitution](https://newfivefour.com/golang-regex-replace-split.html). In fact, all `matchesUrl`
strings are evaluated using regular expressions. If no brackets are given (e.g. `http://localhost/`) the field will
be transformed to `^http://localhost/$`.

Let's come back to our articles example. The regular expression template `http://localhost/articles/<[0-9]+>`
will yield ^http://localhost/articles/([0-9]+)$. This means that we have exactly one group `([0-9]+)` which we can use for
substitution. So let's use the first group in our rule definition:

```json
{
  "id": "policy-rule",
  "matchesMethods": ["GET"],
  "matchesUrl": "http://localhost/articles/<[0-9]+>",
  "mode": "policy",
  "requiredAction": "get",
  "requiredResource": "resources:mydomain.com:articles:$1"
}
```

Now when a user hits, for example `http://localhost/articles/01234`, ORY Oathkeeper will ask the policy decision point
(ORY Hydra's Warden API) a question similar to this one:

```
{
    "subject": "subjects:mydomain.com:alexa",
    "resource": "resources:mydomain.com:articles:01234",
    "action": "get"
}
```

## Examples

The following examples should give you an idea of how rules work. You can save them as files and import them to oathkeeper
using `oathkeeper import --endpoint http://localhost:4456/ <path/to/file.json>`

### Bypass

```json
{
  "id": "bypass-rule",
  "matchesMethods": ["POST", "GET"],
  "matchesUrl": "http://localhost:4455/api/resource",
  "mode": "bypass"
}
```

### Anonymous

```json
{
  "id": "anonymous-rule",
  "matchesMethods": ["POST", "GET"],
  "matchesUrl": "http://localhost:4455/api/resource",
  "mode": "anonymous"
}
```

### Authenticated

Note that `requiredScopes` is an optional field. If you just wish to require user agents to provide a valid access token,
without any scopes, you can leave out the `requiredScopes` field.

```json
{
  "id": "authenticated-rule",
  "matchesMethods": ["POST", "GET"],
  "matchesUrl": "http://localhost:4455/api/resource",
  "mode": "authenticated",
  "requiredScopes": ["scope.a", "scope.b"]
}
```

### Policy Based Access Control

```json
{
  "id": "policy-rule",
  "matchesMethods": ["POST", "GET"],
  "matchesUrl": "http://localhost:4455/api/resource",
  "mode": "policy",
  "requiredScopes": ["scope.a", "scope.b"],
  "requiredAction": "get",
  "requiredResource": "resources:mydomain.com:some:resource"
}
```

Assuming that you want to grant user `subjects:mydomain.com:alexa` access to that endpoint, you would need to create
a policy like the following one:

```json
{
  "id": "some-policy-id",
  "subjects": [
    "subjects:mydomain.com:alexa"
  ],
  "effect": "allow",
  "resources": [
    "resources:mydomain.com:some:resource"
  ],
  "actions": [
    "get"
  ]
}
```

## Rules REST API

For more information on available fields and exemplary payloads of rules, as well as rule management using HTTP
please refer to the [REST API docs](https://oathkeeper.docs.apiary.io/#)

## Rules CLI API

Management of rules is not only possible through the REST API, but additionally using the ORY Oathkeeper CLI.
For help on how to manage the CLI, type `oathkeeper help rules`.
