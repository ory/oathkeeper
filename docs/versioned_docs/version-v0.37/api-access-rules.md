---
id: api-access-rules
title: API Access Rules
---

ORY Oathkeeper reaches decisions to allow or deny access by applying Access
Rules. Access Rules can be stored on the file system, set as an environment
variable, or fetched from HTTP(s) remotes. These repositories can be configured
in the configuration file (`oathkeeper -c ./path/to/config.yml ...`)

```yaml
# Configures Access Rules
access_rules:
  # Locations (list of URLs) where access rules should be fetched from on boot.
  # It is expected that the documents at those locations return a JSON or YAML Array containing ORY Oathkeeper Access Rules.
  repositories:
    # If the URL Scheme is `file://`, the access rules (an array of access rules is expected) will be
    # fetched from the local file system.
    - file://path/to/rules.json
    # If the URL Scheme is `inline://`, the access rules (an array of access rules is expected)
    # are expected to be a base64 encoded (with padding!) JSON/YAML string (base64_encode(`[{"id":"foo-rule","authenticators":[....]}]`)):
    - inline://W3siaWQiOiJmb28tcnVsZSIsImF1dGhlbnRpY2F0b3JzIjpbXX1d
    # If the URL Scheme is `http://` or `https://`, the access rules (an array of access rules is expected) will be
    # fetched from the provided HTTP(s) location.
    - https://path-to-my-rules/rules.json
  # Determines a matching strategy for the access rules . Currently supported values are `glob` and `regexp`. Empty string defaults to regexp.
  matching_strategy: glob
```

or by setting the equivalent environment variable:

```bash
$ export ACCESS_RULES_REPOSITORIES='file://path/to/rules.json,https://path-to-my-rules/rules.json,inline://W3siaWQiOiJmb28tcnVsZSIsImF1dGhlbnRpY2F0b3JzIjpbXX1d'
```

The repository (file, inline, remote) must be formatted either as a JSON or a
YAML array containing the access rules:

```shell
$ cat ./rules.json
[{
    "id": "my-first-rule"
},{
    "id": "my-second-rule"
}]

$ cat ./rules.yaml
- id: my-first-rule
  version: v0.36.0-beta.4
  authenticators:
    - handler: noop
- id: my-second-rule
  version: v0.36.0-beta.4
  authorizer:
    handler: allow
```

## Access Rule Format

Access Rules have four principal keys:

- `id` (string): The unique ID of the Access Rule.
- `version` (string): The version of ORY Oathkeeper this rule targets with out
  the `+oryOS.<x>` appendix. ORY Oathkeeper is able to migrate access rules
  across versions. If left empty ORY Oathkeeper will assume that the rule is
  using the same tag as the version that is running.
- `upstream` (object): The location of the server where requests matching this
  rule should be forwarded to. This only needs to be set when using the ORY
  Oathkeeper Proxy as the Decision API does not forward the request to the
  upstream.
  - `url` (string): The URL the request will be forwarded to.
  - `preserve_host` (bool): If set to `false` (default), the forwarded request
    will include the host and port of the `url` value. If `true`, the host and
    port of the ORY Oathkeeper Proxy will be used instead:
    - `false`: Incoming HTTP Header `Host: mydomain.com`-> Forwarding HTTP
      Header `Host: someservice.intranet.mydomain.com:1234`
  - `strip_path` (string): If set, replaces the provided path prefix when
    forwarding the requested URL to the upstream URL:
    - set to `/api/v1`: Incoming HTTP Request at `/api/v1/users` -> Forwarding
      HTTP Request at `/users`.
    - unset: Incoming HTTP Request at `/api/v1/users` -> Forwarding HTTP Request
      at `/api/v1/users`.
- `match` (object): Defines the URL(s) this Access Rule should match.
  - `methods` (string[]): Array of HTTP methods (e.g. GET, POST, PUT, DELETE,
    ...).
  - `url` (string): The URL that should be matched. You can use regular
    expressions or glob patterns in this field to match more than one url. The
    matching strategy (glob or regexp) is defined in the global configuration
    file as `access_rules.matching_strategy`. This matcher ignores query
    parameters. Regular expressions (or glob patterns) are encapsulated in
    brackets `<` and `>`.

    Regular expressions examples:
    - `https://mydomain.com/` matches `https://mydomain.com/` and does not match
      `https://mydomain.com/foo` or `https://mydomain.com`.
    - `<https|http>://mydomain.com/<.*>` matches:`https://mydomain.com/` or
      `http://mydomain.com/foo`. Does not match: `https://other-domain.com/` or
      `https://mydomain.com`.
    - `http://mydomain.com/<[[:digit:]]+>` matches `http://mydomain.com/123` and
      does not match `http://mydomain/abc`.
    - `http://mydomain.com/<(?!protected).*>` matches
      `http://mydomain.com/resource` and does not match
      `http://mydomain.com/protected`

    [Glob](http://tldp.org/LDP/GNU-Linux-Tools-Summary/html/x11655.htm)
      patterns examples:
    - `https://mydomain.com/<m?n>` matches `https://mydomain.com/man` and does
      not match `http://mydomain.com/foo`.
    - `https://mydomain.com/<{foo*,bar*}>` matches `https://mydomain.com/foo` or
      `https://mydomain.com/bar` and does not match `https://mydomain.com/any`.
- `authenticators`: A list of authentication handlers that authenticate the
  provided credentials. Authenticators are checked iteratively from index `0` to
  `n` and the first authenticator to return a positive result will be the one
  used. If you want the rule to first check a specific authenticator before
  "falling back" to others, have that authenticator as the first item in the
  array. For the full list of available authenticators, click
  [here](pipeline/authn.md).
- `authorizer`: The authorization handler which will try to authorize the
  subject ("user") from the previously validated credentials making the request.
  For example, you could check if the subject ("user") is part of the "admin"
  group or if he/she has permission to perform that action. For the full list of
  available authorizers, click [here](pipeline/authz.md).
- `mutators`: A list of mutation handlers that transform the HTTP request before
  forwarding it. A common use case is generating a new set of credentials (e.g.
  JWT) which then will be forwarded to the upstream server. When using ORY
  Oathkeeper's Decision API, it is expected that the API Gateway forwards the
  mutated HTTP Headers to the upstream server. For the full list of available
  mutators, click [here](pipeline/mutator.md).
- `errors`: A list of error handlers that are executed when any of the previous
  handlers (e.g. authentication) fail. Error handlers define what to do in case
  of an error, for example redirect the user to the login endpoint when a
  unauthorized (HTTP Status Code 401) error occurs. If left unspecified, errors
  will always be handled as JSON responses unless the global configuration key
  `errors.fallback` was changed. For more information on error handlers, click
  [here](pipeline/error.md).

**Examples**

Rule in JSON format:

```json
{
  "id": "some-id",
  "version": "v0.36.0-beta.4",
  "upstream": {
    "url": "http://my-backend-service",
    "preserve_host": true,
    "strip_path": "/api/v1"
  },
  "match": {
    "url": "http://my-app/some-route/<.*>",
    "methods": ["GET", "POST"]
  },
  "authenticators": [{ "handler": "noop" }],
  "authorizer": { "handler": "allow" },
  "mutators": [{ "handler": "noop" }],
  "errors": [{ "handler": "json" }]
}
```

Rule in YAML format:

```yaml
id: some-id
version: v0.36.0-beta.4
upstream:
  url: http://my-backend-service
  preserve_host: true
  strip_path: /api/v1
match:
  url: http://my-app/some-route/<.*>
  methods:
    - GET
    - POST
authenticators:
  - handler: noop
authorizer:
  hander: allow
mutators:
  - handler: noop
errors:
  - handler: json
```

## Handler configuration

Handlers (Authenticators, Mutators, Authorizers, Errors) sometimes require
configuration. The configuration can be defined globally as well as per Access
Rule. The configuration from the Access Rule is overrides values from the global
configuration.

**oathkeeper.yml**

```yaml
authenticators:
  anonymous:
    enabled: true
    config:
      subject: anon
```

**rule.json**

```json
{
  "id": "some-id",
  "upstream": {
    "url": "http://my-backend-service",
    "preserve_host": true,
    "strip_path": "/api/v1"
  },
  "match": {
    "url": "http://my-app/some-route/<.*>",
    "methods": ["GET", "POST"]
  },
  "authenticators": [
    { "handler": "anonymous", "config": { "subject": "anon" } }
  ],
  "authorizer": { "handler": "allow" },
  "mutators": [{ "handler": "noop" }]
}
```

## Scoped Credentials

Some credentials are scoped. For example, OAuth 2.0 Access Tokens usually are
scoped ("OAuth 2.0 Scope"). Scope validation depends on the meaning of the
scope. Therefore, wherever ORY Oathkeeper validates a scope, these scope
strategies are supported:

- `hierarchic`: Scope `foo` matches `foo`, `foo.bar`, `foo.baz` but not `bar`
- `wildcard`: Scope `foo.*` matches `foo`, `foo.bar`, `foo.baz` but not `bar`.
  Scope `foo` matches `foo` but not `foo.bar` nor `bar`
- `exact`: Scope `foo` matches `foo` but not `bar` nor `foo.bar`
- `none`: Scope validation is disabled. If however a scope is configured to be
  validated, the request will fail with an error message.

## Match strategy behavior

With the **Regular expression** strategy, you can use the extracted groups in
all handlers where the substitutions are supported by using the Go
[`text/template`](https://golang.org/pkg/text/template/) package, receiving the
[AuthenticationSession](https://github.com/ory/oathkeeper/blob/master/pipeline/authn/authenticator.go#L39)
struct:

```go
type AuthenticationSession struct {
	Subject      string
	Extra        map[string]interface{}
	Header       http.Header
	MatchContext MatchContext
}

type MatchContext struct {
	RegexpCaptureGroups []string
	URL                 *url.URL
}
```

**Examples**

If the match URL is `<https|http>://mydomain.com/<.*>` and the request is
`http://mydomain.com/foo`, the `MatchContext` field will contain

- `RegexpCaptureGroups`: ["http", "foo"]
- `URL`: "http://mydomain.com/foo"
