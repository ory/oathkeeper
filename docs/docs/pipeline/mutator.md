---
id: mutator
title: Mutators
---

A mutator transforms the credentials from incoming requests to credentials that
your backend understands. For example, the `Authorization: basic` header might
be transformed to `X-User: <subject-id>`. This allows you to write backends that
do not care if the original request was an anonymous one, an OAuth 2.0 Access
Token, or some other credential type. All your backend has to do is understand,
for example, the `X-User:`.

The Access Control Decision API will return the mutated result as the HTTP
Response.

## `noop`

This mutator does not transform the HTTP request and simply forwards the headers
as-is. This is useful if you don't want to replace, for example,
`Authorization: basic` with `X-User: <subject-id>`.

### Configuration

```yaml
# Global configuration file oathkeeper.yml
mutators:
  noop:
    # Set enabled to true if the authenticator should be enabled and false to disable the authenticator. Defaults to false.
    enabled: true
```

```yaml
# Some Access Rule: access-rule-1.yaml
id: access-rule-1
# match: ...
# upstream: ...
mutators:
  - handler: noop
```

### Access Rule Example

```shell
$ cat ./rules.json
{
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
    "handler": "allow"
  },
  "mutators": [
    {
      "handler": "noop"
    }
  ]
}

$ curl -X GET http://my-app/some-route

HTTP/1.0 200 Status OK
The request has been allowed! The original HTTP Request has not been modified.
```

## `id_token`

This mutator takes the authentication information (e.g. subject) and transforms
it to a signed JSON Web Token, and more specifically to an OpenID Connect ID
Token. Your backend can verify the token by fetching the (public) key from the
`/.well-known/jwks.json` endpoint provided by the ORY Oathkeeper API.

Let's say a request is made to a resource protected by ORY Oathkeeper using
Basic Authorization:

```
GET /api/resource HTTP/1.1
Host: www.example.com
Authorization: Basic Zm9vOmJhcg==
```

Assuming that ORY Oathkeeper is granting the access request,
`Basic Zm9vOmJhcg==` will be replaced with a cryptographically signed JSON Web
Token:

```
GET /api/resource HTTP/1.1
Host: internal-api-endpoint-dns
Authorization: Bearer <jwt-signed-id-token>
```

Now, the protected resource is capable of decoding and validating the JSON Web
Token using the public key supplied by ORY Oathkeeper's API. The public key for
decoding the ID token is available at ORY Oathkeeper's `/.well-known/jwks.json`
endpoint:

```
http://oathkeeper:4456/.well-known/jwks.json
```

The related flow diagram looks like this:

[![ID Token Transformation](https://mermaid.ink/img/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgcGFydGljaXBhbnQgQyBhcyBDbGllbnRcbiAgICBwYXJ0aWNpcGFudCBPIGFzIE9hdGhrZWVwZXIgUHJveHlcbiAgICBwYXJ0aWNpcGFudCBBIGFzIFByb3RlY3RlZCBTZXJ2ZXIvQVBJXG4gICAgQy0-Pk86IEF1dGhvcml6YXRpb246IEJhc2ljIC4uLi5cbiAgICBPLS0-Pk86IFZhbGlkYXRlIGNyZWRlbnRpYWxzXG4gICAgTy0-PkE6IEF1dGhvcml6YXRpb246IEJlYXJlciBKLlcuVFxuICAgIEEtLT4-TzogRmV0Y2ggUHVibGljIEtleVxuICAgIEEtLT4-QTogVmFsaWRhdGUgSldUIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQiLCJ0aGVtZUNTUyI6Ii5sYWJlbCBmb3JlaWduT2JqZWN0IHsgb3ZlcmZsb3c6IHZpc2libGU7IGZvbnQtc2l6ZTogMTNweCB9In19)](https://mermaid-js.github.io/mermaid-live-editor/#/edit/eyJjb2RlIjoic2VxdWVuY2VEaWFncmFtXG4gICAgcGFydGljaXBhbnQgQyBhcyBDbGllbnRcbiAgICBwYXJ0aWNpcGFudCBPIGFzIE9hdGhrZWVwZXIgUHJveHlcbiAgICBwYXJ0aWNpcGFudCBBIGFzIFByb3RlY3RlZCBTZXJ2ZXIvQVBJXG4gICAgQy0-Pk86IEF1dGhvcml6YXRpb246IEJhc2ljIC4uLi5cbiAgICBPLS0-Pk86IFZhbGlkYXRlIGNyZWRlbnRpYWxzXG4gICAgTy0-PkE6IEF1dGhvcml6YXRpb246IEJlYXJlciBKLlcuVFxuICAgIEEtLT4-TzogRmV0Y2ggUHVibGljIEtleVxuICAgIEEtLT4-QTogVmFsaWRhdGUgSldUIiwibWVybWFpZCI6eyJ0aGVtZSI6ImRlZmF1bHQiLCJ0aGVtZUNTUyI6Ii5sYWJlbCBmb3JlaWduT2JqZWN0IHsgb3ZlcmZsb3c6IHZpc2libGU7IGZvbnQtc2l6ZTogMTNweCB9In19)

Let's say the `oauth2_client_credentials` authenticator successfully
authenticated the credentials `client-id:client-secret`. This mutator will craft
an ID Token (JWT) with the following exemplary claims:

```json
{
  "iss": "https://server.example.com",
  "sub": "client-id",
  "aud": "s6BhdRkqt3",
  "jti": "n-0S6_WzA2Mj",
  "exp": 1311281970,
  "iat": 1311280970
}
```

The ID Token Claims are as follows:

- `iss`: Issuer Identifier for the Issuer of the response. The iss value is a
  case sensitive URL using the https scheme that contains scheme, host, and
  optionally, port number and path components and no query or fragment
  components. Typically, this is the URL of ORY Oathkeeper, for example:
  `https://oathkeeper.myapi.com`.
- `sub`: Subject Identifier. A locally unique and never reassigned identifier
  within the Issuer for the End-User, which is intended to be consumed by the
  Client, e.g., 24400320 or AItOawmwtWwcT0k51BayewNvutrJUqsvl6qs7A4. It must not
  exceed 255 ASCII characters in length. The sub value is a case sensitive
  string. The End-User might also be an OAuth 2.0 Client, given that the access
  token was granted using the OAuth 2.0 Client Credentials flow.
- `aud`: Audience(s) that this ID Token is intended for. It MUST contain the
  OAuth 2.0 client_id of the Relying Party as an audience value. It MAY also
  contain identifiers for other audiences. In the general case, the aud value is
  an array of case sensitive strings.
- `exp`: Expiration time on or after which the ID Token MUST NOT be accepted for
  processing. The processing of this parameter requires that the current
  date/time MUST be before the expiration date/time listed in the value. Its
  value is a JSON number representing the number of seconds from
  1970-01-01T0:0:0Z as measured in UTC until the date/time. See RFC 3339
  [RFC3339] for details regarding date/times in general and UTC in particular.
- `iat`: Time at which the JWT was issued. Its value is a JSON number
  representing the number of seconds from 1970-01-01T0:0:0Z as measured in UTC
  until the date/time.
- `jti`: A cryptographically strong random identifier to ensure the ID Token's
  uniqueness.

### Global Configuration

### Configuration

- `issuer_url` (string, required) - Sets the "iss" value of the ID Token.
- `jwks_url` (string, required) - Sets the URL where keys should be fetched
  from. Supports remote locations (http, https) as well as local filesystem
  paths.
- `ttl` (string, optional) - Sets the time-to-live of the ID token. Defaults to
  one minute. Valid time units are: s (second), m (minute), h (hour).
- `claims` (string, optional) - Allows you to customize the ID Token claims and
  support Go Templates. For more information, check section [Claims](#claims)

```yaml
# Global configuration file oathkeeper.yml
mutators:
  id_token:
    # Set enabled to true if the authenticator should be enabled and false to disable the authenticator. Defaults to false.
    enabled: true
    config:
      issuer_url: https://my-oathkeeper/
      jwks_url: https://fetch-keys/from/this/location.json
      # jwks_url: file:///from/this/absolute/location.json
      # jwks_url: file://../from/this/relative/location.json
      ttl: 60s
      claims:
        '{"aud": ["https://my-backend-service/some/endpoint"],"def": "{{ print
        .Extra.some.arbitrary.data }}"}'
```

```yaml
# Some Access Rule: access-rule-1.yaml
id: access-rule-1
# match: ...
# upstream: ...
mutators:
  - handler: id_token
    config:
      issuer_url: https://my-oathkeeper/
      jwks_url: https://fetch-keys/from/this/location.json
      # jwks_url: file:///from/this/absolute/location.json
      # jwks_url: file://../from/this/relative/location.json
      ttl: 60s
      claims:
        '{"aud": ["https://my-backend-service/some/endpoint"],"def": "{{ print
        .Extra.some.arbitrary.data }}"}'
```

The first private key found in the JSON Web Key Set defined by
`mutators.id_token.jwks_url` will be used for signing the JWT:

- If the first key found is a symmetric key (`HS256` algorithm), that key will
  be used. That key **will not** be broadcasted at `/.well-known/jwks.json`. You
  must manually configure the upstream to be able to fetch the key (e.g. from an
  environment variable).
- If the first key found is an asymmetric private key (e.g. `RS256`, `ES256`,
  ...), that key will be used. The related public key will be broadcasted at
  `/.well-known/jwks.json`.

#### Claims

This mutator allows you to specify custom claims, like the audience of ID
tokens, via the `claims` field of the mutator's `config` field. The keys
represent names of claims and the values are arbitrary data structures which
will be parsed by the Go [text/template](https://golang.org/pkg/text/template/)
package for value substitution, receiving the `AuthenticationSession` struct.

For more details please check [Session variables](index.md#session)

The claims configuration expects a string which is expected to be valid JSON:

```json
{
  "handler": "id_token",
  "config": {
    "claims": "{\"aud\": [\"https://my-backend-service/some/endpoint\"],\"def\": \"{{ print .Extra.some.arbitrary.data }}\"}"
  }
}
```

Please keep in mind that certain keys (such as the `sub`) claim **can not** be
overwritten!

### Access Rule Example

```shell
$ cat ./rules.json
{
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
    "handler": "allow"
  },
  "mutators": [
    {
      "handler": "id_token",
      "config": {
        "aud": [
          "audience-1",
          "audience-2"
        ],
        "claims": "{\"abc\": \"{{ print .Subject }}\",\"def\": \"{{ print .Extra.some.arbitrary.data }}\"}"
      }
    }
  ]
}
```

## `header`

This mutator will transform the request, allowing you to pass the credentials to
the upstream application via the headers. This will augment, for example,
`Authorization: basic` with `X-User: <subject-id>`.

### Configuration

- `headers` (object (`string: string`), required) - A keyed object
  (`string:string`) representing the headers to be added to this request, see
  section [headers](#headers).

```yaml
# Global configuration file oathkeeper.yml
mutators:
  header:
    # Set enabled to true if the authenticator should be enabled and false to disable the authenticator. Defaults to false.
    enabled: true
    config:
      headers:
        X-User: '{{ print .Subject }}'
        X-Some-Arbitrary-Data: '{{ print .Extra.some.arbitrary.data }}'
```

```yaml
# Some Access Rule: access-rule-1.yaml
id: access-rule-1
# match: ...
# upstream: ...
mutators:
  - handler: header
    config:
      headers:
        X-User: '{{ print .Subject }}'
        X-Some-Arbitrary-Data: '{{ print .Extra.some.arbitrary.data }}'
```

#### Headers

The headers are specified via the `headers` field of the mutator's `config`
field. The keys are the header name and the values are a string which will be
parsed by the Go [`text/template`](https://golang.org/pkg/text/template/)
package for value substitution, receiving the `AuthenticationSession` struct.

For more details please check [Session variables](index.md#session)

### Access Rule Example

```json
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
    "handler": "allow"
  },
  "mutators": [
    {
      "handler": "header",
      "config": {
        "headers": {
          "X-User": "{{ print .Subject }}",
          "X-Some-Arbitrary-Data": "{{ print .Extra.some.arbitrary.data }}"
        }
      }
    }
  ]
}
```

## `cookie`

This mutator will transform the request, allowing you to pass the credentials to
the upstream application via the cookies.

### Configuration

- `cookies` (object (`string: string`), required) - A keyed object
  (`string:string`) representing the cookies to be added to this request, see
  section [cookies](#cookies).

```yaml
# Global configuration file oathkeeper.yml
mutators:
  cookie:
    # Set enabled to true if the authenticator should be enabled and false to disable the authenticator. Defaults to false.
    enabled: true
    config:
      cookies:
        user: "{{ print .Subject }}",
        some-arbitrary-data: "{{ print .Extra.some.arbitrary.data }}"
```

```yaml
# Some Access Rule: access-rule-1.yaml
id: access-rule-1
# match: ...
# upstream: ...
mutators:
  - handler: cookie
    config:
      cookies:
        user: "{{ print .Subject }}",
        some-arbitrary-data: "{{ print .Extra.some.arbitrary.data }}"
```

### Cookies

The cookies are specified via the `cookies` field of the mutators `config`
field. The keys are the cookie name and the values are a string which will be
parsed by the Go [`text/template`](https://golang.org/pkg/text/template/)
package for value substitution, receiving the `AuthenticationSession` struct.

For more details please check [Session variables](index.md#session)

##### Example

```json
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
    "handler": "allow"
  },
  "mutators": [
    {
      "handler": "cookie",
      "config": {
        "cookies": {
          "user": "{{ print .Subject }}",
          "some-arbitrary-data": "{{ print .Extra.some.arbitrary.data }}"
        }
      }
    }
  ]
}
```

## `hydrator`

This mutator allows for fetching additional data from external APIs, which can
be then used by other mutators. It works by making an upstream HTTP call to an
API specified in the **Per-Rule Configuration** section below. The request is a
POST request and it contains JSON representation of
[AuthenticationSession](https://github.com/ory/oathkeeper/blob/master/pipeline/authn/authenticator.go#L39)
struct in body, which is:

```json
{
  "subject": String,
  "extra": Object,
  "header": Object,
  "match_context": {
    "regexp_capture_groups": Object,
    "url": Object
  }
}
```

As a response the mutator expects similiar JSON object, but with `extra` or
`header` fields modified.

Example request/response payload:

```json
{
  "subject": "anonymous",
  "extra": {
    "foo": "bar"
  },
  "header": {
    "foo": ["bar1", "bar2"]
  },
  "match_context": {
    "regexp_capture_groups": ["http", "foo"],
    "url": "http://domain.com/foo"
  }
}
```

The AuthenticationSession from this object replaces the original one and is
passed to the next mutator, where it can be used to e.g. set a particular cookie
to the value received from an API.

Setting `extra` field does not transform the HTTP request, whereas headers set
in the `header` field will be added to the final request headers.

### Cache

This handler supports caching. If caching is enabled, the `api.url` configuration value
and the the full `AuthenticationSession` payload.

:::info

Because the cache key is quite complex, the caching handler has a higher chance of cache misses.
This will be improved in future versions.

:::

### Configuration

- `api.url` (string - required) - The API URL.
- `api.auth.basic.*` (optional) - Enables HTTP Basic Authorization.
- `api.auth.retry.*` (optional) - Configures the retry logic.
- `cache.ttl` (optional) - Configures how long to cache hydrate requests

```yaml
# Global configuration file oathkeeper.yml
mutators:
  hydrator:
    # Set enabled to true if the authenticator should be enabled and false to disable the authenticator. Defaults to false.
    enabled: true
    config:
      api:
        url: http://my-backend-api
        auth:
          basic:
            username: someUserName
            password: somePassword
        retry:
          give_up_after: 2s
          max_delay: 100ms
      cache:
        ttl: 60s
```

```yaml
# Some Access Rule: access-rule-1.yaml
id: access-rule-1
# match: ...
# upstream: ...
mutators:
  - handler: hydrator
    config:
      api:
        url: http://my-backend-api
        auth:
          basic:
            username: someUserName
            password: somePassword
        retry:
          give_up_after: 2s
          max_delay: 100ms
      cache:
        ttl: 60s
```

### Access Rule Example

```json
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
    "handler": "allow"
  },
  "mutators": [
    {
      "handler": "hydrator",
      "config": {
        "api": {
          "url": "http://my-backend-api"
        }
      }
    },
    {
      "handler": "cookie",
      "config": {
        "cookies": {
          "some-arbitrary-data": "{{ print .Extra.cookie }}"
        }
      }
    }
  ]
}
```
