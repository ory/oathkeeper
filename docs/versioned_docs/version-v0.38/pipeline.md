---
id: pipeline
title: Access Rule Pipeline
---

Read more about the
[principal components and execution pipeline of access rules](api-access-rules.md)
if you have not already. This chapter explains the different pipeline handlers
available to you:

- [Authentication handlers](pipeline/authn.md) inspect HTTP requests (e.g. the
  HTTP Authorization Header) and execute some business logic that return true
  (for authentication ok) or false (for authentication invalid) as well as a
  subject ("user"). The subject is typically the "user" that made the request,
  but it could also be a machine (if you have machine-2-machine interaction) or
  something different.
- [Authorization handlers](pipeline/authz.md): ensure that a subject ("user")
  has the right permissions. For example, a specific endpoint might only be
  accessible to subjects ("users") from group "admin". The authorizer handles
  that logic.
- [Mutation handlers](pipeline/mutator.md): transforms the credentials from
  incoming requests to credentials that your backend understands. For example,
  the `Authorization: basic` header might be transformed to
  `X-User: <subject-id>`. This allows you to write backends that do not care if
  the original request was an anonymous one, an OAuth 2.0 Access Token, or some
  other credential type. All your backend has to do is understand, for example,
  the `X-User:`.
- [Error handlers](pipeline/error.md): are responsible for executing logic
  after, for example, authentication or authorization failed. ORY Oathkeeper
  supports different error handlers and we will add more as the project
  progresses.

## Templating

Some handlers such as the [ID Token Mutator](pipeline/mutator.md#id_token)
support templating using
[Golang Text Templates](https://golang.org/pkg/text/template/)
([examples](https://blog.gopheracademy.com/advent-2017/using-go-templates/)).
The [sprig](http://masterminds.github.io/sprig/) is also supported, on top of
these two functions:

```go
var _ = template.FuncMap{
    "print": func(i interface{}) string {
        if i == nil {
            return ""
        }
        return fmt.Sprintf("%v", i)
    },
    "printIndex": func(element interface{}, i int) string {
        if element == nil {
            return ""
        }

        list := reflect.ValueOf(element)

        if list.Kind() == reflect.Slice && i < list.Len() {
            return fmt.Sprintf("%v", list.Index(i))
        }

        return ""
    },
}
```

## Session

In all configurations supporting [templating](#templating) instructions, it's
possible to use the
[`AuthenticationSession`](https://github.com/ory/oathkeeper/blob/master/pipeline/authn/authenticator.go#L39)
struct content.

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
	Method              string
	Header              http.Header
}
```

### RegexpCaptureGroups

### Configuration Examples

To use the subject extract to the token

```json
{ "config_field": "{{ print .subject }}" }
```

To use any arbitrary header value from the request headers

```json
{ "config_field": "{{ .MatchContext.Header.Get \"some_header\" }}" }
```

To use an embedded value in the `Extra` map (most of the time, it's a JWT token
claim)

```json
{ "config_field": "{{ print .Extra.some.arbitrary.data }}" }
```

To use a Regex capture from the request URL Note the usage of `printIndex` to
print a value from the array

```json
{
  "claims": "{\"aud\": \"{{ print .Extra.aud }}\", \"resource\": \"{{ printIndex .MatchContext.RegexpCaptureGroups 0 }}\""
}
```

To display a string array to JSON format, we can use the
[fmt printf](https://golang.org/pkg/fmt/) function

```json
{
  "claims": "{\"aud\": \"{{ print .Extra.aud }}\", \"scope\": {{ printf \"%+q\" .Extra.scp }}}"
}
```

Note that the `AuthenticationSession` struct has a field named `Extra` which is
a `map[string]interface{}`, which receives varying introspection data from the
authentication process. Because the contents of `Extra` are so variable, nested
and potentially non-existent values need special handling by the `text/template`
parser, and a `print` FuncMap function has been provided to ensure that
non-existent map values will simply return an empty string, rather than
`<no value>`.

If you find that your field contain the string `<no value>` then you have most
likely omitted the `print` function, and it is recommended you use it for all
values out of an abundance of caution and for consistency.

In the same way, a `printIndex` FuncMap function is provided to avoid _out of
range_ exception to access in a array. It can be useful for the regexp captures
which depend of the request.
