---
id: error
title: Error Handlers
---

A error handler is responsible for executing logic after, for example,
authentication or authorization failed. ORY Oathkeeper supports different error
handlers and we will add more as the project progresses.

A error handler can be configured to match on certain conditions, for example,
it is possible to configure the `json` error handler to only be executed if the
HTTP Header `Accept` contains `application/json`.

Each error handler has two keys:

- `handler` (string, required): Defines the handler (e.g. `noop`) to be used.
- `config` (object, optional): Configures the handler. Configuration keys vary
  per handler. The configuration can be defined in the global configuration
  file, or per access rule.

**Example**

```json
{
  "errors": [
    {
      "handler": "json",
      "config": {}
    }
  ]
}
```

You can define more than one error handler in the Access Rule. Depending on
their matching conditions (see next chapter), the appropriate error handler will
be chosen.

Please be aware that defining error handlers with overlapping matching
conditions will cause errors, because ORY Oathkeeper will not know which error
handler to execute!

## Error Matching

You can configure the error handlers in such a way, that - for example - ORY
Oathkeeper responds, in the case of an error, with

- a JSON response, such as
  `{"error":{"code":403,"status":"Forbidden","message":"Access credentials are not sufficient to access this resource"}}`,
  when the client that expects JSON (`Accept: application/json`);
- an XML response when the API Client expects XML (`Accept: application/xml`);
- a HTTP Redirect (HTTP Status Found - 302) to `/login` when the endpoint is
  directly (no AJAX) accessed from a browser
  (`Accept: text/html,application/xhtml+xml`).

There are also other possible matching strategies - such as defining a response
per error type (unauthorized, forbidden, internal_server_error, ...) , per HTTP
`Content-Type` Header (similar to `Accept`), or based on the Remote IP Address.

All match definitions are set in the handler's config, using the `when` key.
This is the same for all handlers!

```json5
{
  handler: 'json', // or redirect, www_authenticate, ...
  config: {
    when: [
      {
        error: ['unauthorized', '...', '...'],
      },
    ],
  },
}
```

If `when` is empty, then no conditions are applied and the error handler is
always matching! In fact, this is also true for all subkeys. If left empty, the
matching condition will not be applied and is thus always true!

### Fallback

Error handling can be set globally and per access rule. ORY Oathkeeper will
first check for any access rule specific error handling before falling back to
the globally defined error handling.

Similar to other pipeline handlers (authentication, authorization, mutation),
you must enable the error handlers in the global ORY Oathkeeper config, except
for the `json` error handler which is always enabled by default:

```yaml
# .oathkeeper.yaml
errors:
  handlers:
    json:
      enabled: true # this is true by default
      # config:
      #   when: ...
    redirect:
      enabled: true # this is false by default
      # config:
      #   when: ...
```

As discussed in the previous section, when `config.when` is empty, the error
handler will always match. This of course is a problem because ORY Oathkeeper
now does not know if it should redirect or send a JSON error!

Therefore, an additional configuration - called `fallback` - is available:

```yaml
# .oathkeeper.yaml
errors:
  # `["json"]` is the default!
  fallback:
    - json

  handlers:
    json:
      enabled: true # this is true by default
      # config:
      #   when: ...
    redirect:
      enabled: true # this is false by default
      config:
        to: http://mywebsite/login
      # when: ...
```

This feature tells ORY Oathkeeper that the `json` error handler should be used
as fallback. You could also define multiple fallback handlers - the first
matching handler will be the one and only executed! This makes sense if you
additionally configure the `when` section:

```yaml
# .oathkeeper.yaml
errors:
  fallback:
    - redirect
    - json

  handlers:
    json:
      enabled: true
    redirect:
      enabled: true
      config:
        when:
          - request:
              header:
                accept:
                  - text/*
```

In this configuration example, ORY Oathkeeper would first check if the HTTP
Request Header contains `Accept: text/html` (or `text/xhtml`, `text/text`, ...)
and if not, would return a JSON Error Message.

### Matchers

All matchers are defined under the `config.when` key of the error handler, both
in the global config and in the access rule:

```json5
// access-rule.json
{
  handler: 'json',
  config: {
    when: [
      {
        error: ['unauthorized', '...', '...'],
      },
    ],
  },
}
```

```yaml
# .oathkeeper.yaml
errors:
  handlers:
    redirect:
      enabled: true
      config:
        when:
          - error:
              - unauthorized
              - authentication_handler_no_match
              - ...
              - ...
```

You can define multiple when clauses which allows you to differentiate between
error types and HTTP Requests. The when sections are combined with `OR` while
the subkeys (`error`, `request.header.accept`, `request.header.content_type`,
...) are matched with `AND`. Keys that have arrays as values (`error`,
`request.header.accept`, `request.header.content_type`, ...) are usually matched
with `OR`:

```yaml
# .oathkeeper.yaml
errors:
  handlers:
    redirect:
      enabled: true
      config:
        when:
          - error:
              - unauthorized
              # OR
              - internal_server_error

            # AND
            request:
              remote_ip:
                match:
                  - 192.168.1.0/24
                  # OR
                  - 192.178.1.0/24

          # OR
          - error:
              - forbidden
              # OR
              - not_found

            # AND
            request:
              header:
                accept:
                  - text/html
                  # OR
                  - text/xhtml

                # AND
                content_type:
                  - application/x-www-form-urlencoded
                  # OR
                  - multipart/form-data
```

#### Error

The `config.when.#.error` key may contain zero, one, or multiple error names
that must match for this matching condition to be true. The error names are
derived (lowercase and whitespaces replaced with `_`) from the well-defined
[HTTP Status](https://en.wikipedia.org/wiki/List_of_HTTP_status_codes) messages
such as `Not Found`, `Forbidden`, `Internal Server Error`, and so on.

Here are some examples:

- `Internal Server Error` (500) -> `{"errors": ["internal_server_error"]}`
- `Forbidden` (403) -> `{"errors": ["forbidden"]}`
- `Not Found` (404) -> `{"errors": ["not_found"]}`
- `Bad Request` (400) -> `{"errors": ["bad_request"]}`

Keep in mind that these errors must be emitted by ORY Oathkeeper itself, not by
the upstream API. Therefore, most HTTP Status Codes will not have any effect
because ORY Oathkeeper - as of now - mostly returns 401, 403, 500 error codes.

As discussed previously, if this configuration key is left empty, then all error
types will match!

#### HTTP Request: Remote IP

The HTTP Remote IP is the IP of the Client that initially made the request. The
Remote Address is matched using
[CIDR Notation](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing):

```yaml
config:
  when:
    - request:
        remote_ip:
          match:
            - 192.168.1.0/24
```

This configuration would match a HTTP Request coming directly from
`192.168.1.1`, `192.168.1.2`, and so on.

If ORY Oathkeeper runs behind a Load Balancer or any other type of Reverse
Proxy, you can configure ORY Oathkeeper to check the
[`X-Forwarded-For` HTTP Header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For)
header as well:

```yaml
config:
  when:
    - request:
        remote_ip:
          respect_forwarded_for_header: true # defaults to false
          match:
            - 192.168.1.0/24
```

As discussed previously, if this configuration key is left empty, then all
remote IPs will match!

HTTP Requests that include one of the matching IP Addresses in the
`X-Forwaded-For` HTTP Header, for example
`X-Forwarded-For: 123.123.123.123, ..., 192.168.1.1, ...`, now match this error
handler.

#### HTTP Request Header: Accept

The HTTP `Accept` Header is the most common way to tell an HTTP API what MIME
content type is expected. For example, FireFox sends
`Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8` for
all regular requests (e.g. when opening [www.ory.sh](https://www.ory.sh/)). And
a REST API Client usually sends `Accept: application/json`.

Therefore, using the `Accept` header is one of the most common ways to
distinguish between "regular" browser traffic, REST API traffic, and other types
of HTTP traffic.

In ORY Oathkeeper, you can specify the matching conditions for the Accept header
as follows:

```yaml
config:
  when:
    - request:
        header:
          accept:
            - text/html
            - text/*
```

The defined matching condition would apply if a client sends one of the
following `Accept` headers:

- `Accept: text/html`
- `Accept: text/xhtml`
- `Accept: text/xhtml+xml`
- `Accept: text/...`
- `Accept: text/*`

Most browsers (see the FireFox example) also send wildcard `Accept` headers such
as `*/*`. To prevent multiple conditions to match, HTTP Accept Headers from the
client are interpreted literally, meaning that wildcards are not interpreted.

Assuming the client sends `Accept: */*` and the error condition is set to
`accept: ["text/text"]`, the error condition would not match. If however the
client sends `Accept: text/text` and the error condition is set to
`accept: ["*/*"]`, then the condition would match.

To match against wildcards in the `Accept` header, you have to explicitly define
them in the error condition. Setting the configuration to `accept: ["*/*"]` will
match `Accept: */*` and of course any other type such as `Accept: text/*`
`Accept: text/html`, and so on.

As discussed previously, if this configuration key is left empty, then all
`Accept` headers will match!

#### HTTP Request Header: Content-Type

The
[HTTP Content Type](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Type)
matcher works similar to the `Accept` header. The HTTP Content Type Header
however is much less common, as it is only used in POST, PUT, PATCH requests (or
any other requests that send a HTTP Body).

The main difference however is that the client never (unless it sends malformed
data) sends wildcard MIME types, as the MIME type needs to be deterministic.
It's typically something like `multipart/form-data`,
`application/x-www-form-urlencoded`, or `application/json`.

In ORY Oathkeeper, you can specify the matching conditions for the
`Content-Type` header as follows:

```yaml
config:
  when:
    - request:
        header:
          content_type:
            - multipart/form-data
            # OR
            - application/x-www-form-urlencoded
            # OR
            - application/json
```

As discussed previously, if this configuration key is left empty, then all
`Content-Type` headers will match!

## Error Handlers

### `json`

The `json` Error Handler returns an `application/json` response type. Per
default, error messages are stripped of their details to reduce OSINT attack
surface. You can enable more detailed error messages by setting `verbose` to
`true`. As discussed in the previous section, you can define error matching
conditions under the `when` key.

**Example**

```json5
// access-rule.json
{
  handler: 'json',
  config: {
    verbose: true, // defaults to false
    when: [
      // ...
    ],
  },
}
```

### `redirect`

The `redirect` Error Handler returns a HTTP 302/301 response with a `Location`
Header. As discussed in the previous section, you can define error matching
conditions under the `when` key.

**Example**

```json5
// access-rule.json
{
  handler: 'json',
  config: {
    to: 'http://my-website/login', // required!!
    code: 301, // defaults to 302 - only 301 and 302 are supported.
    when: [
      // ...
    ],
  },
}
```

### `www_authenticate`

The `www_authenticate` Error Handler responds with HTTP 401 and a
[`WWW-Authenticate`](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate)
HTTP Header.

You can configure the `realm` the browser will display. The `realm` is a message
that will be displayed by the browser. Most browsers show a message like "The
website says: `<realm>`". Using a real message is thus more appropriate than a
Realm identifier.

This error handler is "exotic" as WWW-Authenticate is not a common pattern in
today's web. As discussed in the previous section, you can define error matching
conditions under the `when` key.

**Example**

```json5
// access-rule.json
{
  handler: 'json',
  config: {
    realm: 'Please enter your username and password', // Defaults to `Please authenticate.`
    when: [
      // ...
    ],
  },
}
```
