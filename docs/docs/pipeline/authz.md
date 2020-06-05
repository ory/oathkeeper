---
id: authz
title: Authorizers
---

An "authorizer" is responsible for properly permissioning a subject. ORY
Oathkeeper supports different kinds of authorizers. The list of authorizers
increases over time due to new features and requirements.

Authorizers assure that a subject, for instance a "user", has the permissions
necessary to access or perform a particular service. For example, an authorizer
can permit access to an endpoint or URL for specific subjects or "users" from a
specific group "admin". The authorizer permits the subjects the desired access
to the endpoint.

Each authorizer has two keys:

- `handler` (string, required): Defines the handler, e.g. `noop`, to be used.
- `config` (object, optional): Configures the handler. Configuration keys can
  vary for each handler.

**Example**

```json
{
  "authorizer": {
    "handler": "noop",
    "config": {}
  }
}
```

There is a 1:1 mandatory relationship between an authoriser and an access rule.
It is not possible to configure more than one authorizer per Access Rule.

## Authorizer `allow`

This authorizer permits every action allowed.

### Configuration

This handler is not configurable.

To enable this handler, set as follows:

```yaml
# Global configuration file oathkeeper.yml
authorizers:
  allow:
    # Set enabled to "true" to enable the authenticator, and "false" to disable the authenticator. Defaults to "false".
    enabled: true
```

### Access Rule Example

```sh
$ cat ./rules.json

[{
  "id": "some-id",
  "upstream": {
    "url": "http://my-backend-service"
  },
  "match": {
    "url": "http://my-app/some-route",
    "methods": [
      "GET"
    ]
  },
  "authenticators": [{ "handler": "anonymous" }],
  "authorizer": { "handler": "allow" },
  "mutators": [{ "handler": "noop" }]
}]

$ curl -X GET http://my-app/some-route

HTTP/1.0 200 Status OK
The request has been allowed!
```

## Authorizer`deny`

This authorizer considers every action unauthorized therefore "forbidden" or
"disallowed".

### Configuration

This handler is not configurable.

To enable this handler, set:

```yaml
# Global configuration file oathkeeper.yml
authorizers:
  deny:
    # Set enabled to "true" to enable the authenticator, and "false" to disable the authenticator. Defaults to "false".
    enabled: true
```

### Access Rule Example

```sh
$ cat ./rules.json

[{
  "id": "some-id",
  "upstream": {
    "url": "http://my-backend-service"
  },
  "match": {
    "url": "http://my-app/some-route",
    "methods": [
      "GET"
    ]
  },
  "authenticators": [{ "handler": "anonymous" }],
  "authorizer": { "handler": "deny" },
  "mutators": [{ "handler": "noop" }]
}]

$ curl -X GET http://my-app/some-route

HTTP/1.0 403 Forbidden
The request is forbidden!
```

## Authorizer `keto_engine_acp_ory`

This authorizer uses the ORY Keto API to carry out access control using
"ORY-flavored" Access Control Policies. The conventions used in the ORY Keto
project are located on [GitHub ORY Keto](https://github.com/ory/keto) for
consultation prior to using this authorizer.

### Configuration

- `base_url` (string, required) - The base URL of ORY Keto, typically something
  like https://hostname:port/
- `required_action` (string, required) - See section below.
- `required_resource` (string, required) - See section below.
- `subject` (string, optional) - See section below.
- `flavor` (string, optional) - See section below.

#### Resource, Action, Subject

This authorizer has four configuration options, `required_action`,
`required_resource`, `subject`, and `flavor`:

```json
{
  "handler": "keto_engine_acp_ory",
  "config": {
    "required_action": "...",
    "required_resource": "...",
    "subject": "...",
    "flavor": "..."
  }
}
```

All configuration options except `flavor` support Go
[`text/template`](https://golang.org/pkg/text/template/). For example in the
following match configuration:

```json
{
  "match": {
    "url": "http://my-app/api/users/<[0-9]+>/<[a-zA-Z]+>",
    "methods": ["GET"]
  }
}
```

The following example shows how to reference the values matched by or resulting
from the two regular expressions, `<[0-9]+>` and `<[a-zA-Z]+>`. using the
`AuthenticationSession` struct:

```json
{
  "handler": "keto_engine_acp_ory",
  "config": {
    "required_action": "my:action:{{ printIndex .MatchContext.RegexpCaptureGroups 0 }}",
    "required_resource": "my:resource:{{ printIndex .MatchContext.RegexpCaptureGroups 1 }}:foo:{{ printIndex .MatchContext.RegexpCaptureGroups 0 }}"
  }
}
```

Assuming a request to `http://my-api/api/users/1234/foobar` was made, the config
from above would expand to:

```json
{
  "handler": "keto_engine_acp_ory",
  "config": {
    "required_action": "my:action:1234",
    "required_resource": "my:resource:foobar:foo:1234"
  }
}
```

The `subject` field configures the subject that passes to the ORY Keto endpoint.
If `subject` is not specified it will default to
`AuthenticationSession.Subject`.

For more details about supported Go template substitution, see.
[How to use session variables](index.md#session)

#### Example

```yaml
# Global configuration file oathkeeper.yml
authorizers:
  keto_engine_acp_ory:
    # Set enabled to "true" to enable the authenticator, and "false" to disable the authenticator. Defaults to "false".
    enabled: true

    config:
      base_url: http://my-keto/
      required_action: ...
      required_resource: ...
      subject: ...
      flavor: ...
```

```yaml
# Some Access Rule: access-rule-1.yaml
id: access-rule-1
# match: ...
# upstream: ...
authorizers:
  - handler: keto_engine_acp_ory
    config:
      base_url: http://my-keto/
      required_action: ...
      required_resource: ...
      subject: ...
      flavor: ...
```

### Access Rule Example

```shell
$ cat ./rules.json

[{
  "id": "some-id",
  "upstream": {
    "url": "http://my-backend-service"
  },
  "match": {
    "url": "http://my-app/api/users/<[0-9]+>/<[a-zA-Z]+>",
    "methods": [
      "GET"
    ]
  },
  "authenticators": [
    {
      "handler": "anonymous"
    }
  ],
  "authorizer": {
    "handler": "keto_engine_acp_ory",
    "config": {
      "required_action": "my:action:$1",
      "required_resource": "my:resource:$2:foo:$1"
      "subject": "{{ .Extra.email }}",
      "flavor": "exact"
    }
  }
  "mutators": [
    {
      "handler": "noop"
    }
  ]
}]
```

## `remote`

This authorizer performs authorization using a remote authorizer. The authorizer
makes a HTTP POST request to a remote endpoint with the original body request as
body. If the endpoint returns a "200 OK" response code, the access is allowed,
if it returns a "403 Forbidden" response code, the access is denied.

### Configuration

- `remote` (string, required) - The remote authorizer's URL. The remote
  authorizer is expected to return either "200 OK" or "403 Forbidden" to
  allow/deny access.
- `headers` (map of strings, optional) - The HTTP headers sent to the remote
  authorizer. The values will be parsed by the Go
  [`text/template`](https://golang.org/pkg/text/template/) package and applied
  to an
  [`AuthenticationSession`](https://github.com/ory/oathkeeper/blob/master/pipeline/authn/authenticator.go#L40)
  object. See [Session](index.md#session) for more details.

#### Example

```yaml
# Global configuration file oathkeeper.yml
authorizers:
  remote:
    # Set enabled to "true" to enable the authenticator, and "false" to disable the authenticator. Defaults to "false".
    enabled: true

    config:
      remote: http://my-remote-authorizer/authorize
      headers:
        X-Subject: '{{ print .Subject }}'
```

```yaml
# Some Access Rule: access-rule-1.yaml
id: access-rule-1
# match: ...
# upstream: ...
authorizers:
  - handler: remote
    config:
      remote: http://my-remote-authorizer/authorize
      headers:
        X-Subject: '{{ print .Subject }}'
```

### Access Rule Example

```shell
{
  "id": "some-id",
  "upstream": {
    "url": "http://my-backend-service"
  },
  "match": {
    "url": "http://my-app/api/<.*>",
    "methods": ["GET"]
  },
  "authenticators": [
    {
      "handler": "anonymous"
    }
  ],
  "authorizer": {
    "handler": "remote",
    "config": {
      "remote": "http://my-remote-authorizer/authorize",
      "headers": {
        "X-Subject": "{{ print .Subject }}"
      }
    }
  }
  "mutators": [
    {
      "handler": "noop"
    }
  ]
}
```

## `remote_json`

This authorizer performs authorization using a remote authorizer. The authorizer
makes a HTTP POST request to a remote endpoint with a JSON body. If the endpoint
returns a "200 OK" response code, the access is allowed, if it returns a "403
Forbidden" response code, the access is denied.

### Configuration

- `remote` (string, required) - The remote authorizer's URL. The remote
  authorizer is expected to return either "200 OK" or "403 Forbidden" to
  allow/deny access.
- `payload` (string, required) - The request's JSON payload sent to the remote
  authorizer. The string will be parsed by the Go
  [`text/template`](https://golang.org/pkg/text/template/) package and applied
  to an
  [`AuthenticationSession`](https://github.com/ory/oathkeeper/blob/master/pipeline/authn/authenticator.go#L40)
  object. See [Session](index.md#session) for more details.

#### Example

```yaml
# Global configuration file oathkeeper.yml
authorizers:
  remote_json:
    # Set enabled to "true" to enable the authenticator, and "false" to disable the authenticator. Defaults to "false".
    enabled: true

    config:
      remote: http://my-remote-authorizer/authorize
      payload: |
        {
          "subject": "{{ print .Subject }}",
          "resource": "{{ printIndex .MatchContext.RegexpCaptureGroups 0 }}"
        }
```

```yaml
# Some Access Rule: access-rule-1.yaml
id: access-rule-1
# match: ...
# upstream: ...
authorizers:
  - handler: remote_json
    config:
      remote: http://my-remote-authorizer/authorize
      payload: |
        {
          "subject": "{{ print .Subject }}",
          "resource": "{{ printIndex .MatchContext.RegexpCaptureGroups 0 }}"
        }
```

### Access Rule Example

```shell
{
  "id": "some-id",
  "upstream": {
    "url": "http://my-backend-service"
  },
  "match": {
    "url": "http://my-app/api/<.*>",
    "methods": ["GET"]
  },
  "authenticators": [
    {
      "handler": "anonymous"
    }
  ],
  "authorizer": {
    "handler": "remote_json",
    "config": {
      "remote": "http://my-remote-authorizer/authorize",
      "payload": "{\"subject\": \"{{ print .Subject }}\", \"resource\": \"{{ printIndex .MatchContext.RegexpCaptureGroups 0 }}\"}"
    }
  }
  "mutators": [
    {
      "handler": "noop"
    }
  ]
}
```

## `remote_opa`

This authorizer performs authorization against an OPA policy document using the API
of a remote OPA instance. The authorizer makes a HTTP POST request to the OPA instance
with a JSON formatted policy input containing the authenticated subject (if present),
the HTTP method, and path of the request. The OPA API always returns a "200 OK" response
code and a JSON formatted body containing a policy result specifying "allow" as either
true or false.

If the "allow" value of the policy result is true then access is allowed, if it is false
or undefined then access is denied. This authorizer is intended to be a drop-in alternative
to the python middleware in the official Open Policy Agent documentation.

See:
[`OPA HTTP API Authorization Use Case](https://www.openpolicyagent.org/docs/v0.11.0/http-api-authorization/)
[`OPA API Documentation - Get a Document (with Input)](https://www.openpolicyagent.org/docs/latest/rest-api/#get-a-document-with-input)

### Configuration

- `remote` (string, required) - The remote OPA policy document URL located
  under the /v1/data API path.

#### Example

```yaml
# Global configuration file oathkeeper.yml
authorizers:
  remote_json:
    # Set enabled to "true" to enable the authenticator, and "false" to disable the authenticator. Defaults to "false".
    enabled: true

    config:
      remote: http://opa-host:8181/v1/data/example/authz
```

```yaml
# Some Access Rule: access-rule-1.yaml
id: access-rule-1
# match: ...
# upstream: ...
authorizers:
  - handler: remote_json
    config:
      remote: http://opa-host:8181/v1/data/example/authz
```

### Access Rule Example
```shell
{
  "id": "some-id",
  "upstream": {
    "url": "http://my-backend-service"
  },
  "match": {
    "url": "http://my-app/api/<.*>",
    "methods": ["GET"]
  },
  "authenticators": [
    {
      "handler": "anonymous"
    }
  ],
  "authorizer": {
    "handler": "remote_json",
    "config": {
      "remote": "http://opa-host:8181/v1/data/example/authz",
    }
  }
  "mutators": [
    {
      "handler": "noop"
    }
  ]
}

### Generated JSON Payload Example
```shell
{
  "input": {
    "user": "Subject",
    "path": ["request","url","split","into","an","array"]
    "method": "GET"
  }
}
```

### Expected JSON Response Example
```shell
{
  "result": {
    "allow": true
  }
}
```