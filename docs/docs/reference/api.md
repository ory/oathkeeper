---
title: REST API
id: api
---

ORY Oathkeeper is a reverse proxy that checks the HTTP Authorization for
validity against a set of rules. This service uses Hydra to validate access
tokens and policies.

> You are viewing REST API documentation. This documentation is auto-generated
> from a swagger specification which itself is generated from annotations in the
> source code of the project. It is possible that this documentation includes
> bugs and that code samples are incomplete or wrong.
>
> If you find issues in the respective documentation, please do not edit the
> Markdown files directly (as they are generated) but raise an issue on the
> project's GitHub presence instead. This documentation will improve over time
> with your help! If you have ideas how to improve this part of the
> documentation, feel free to share them in a
> [GitHub issue](https://github.com/ory/docs/issues/new) any time.

<a id="ory-oathkeeper-api"></a>

## api

<a id="opIdgetWellKnownJSONWebKeys"></a>

### Lists cryptographic keys

```
GET /.well-known/jwks.json HTTP/1.1
Accept: application/json

```

This endpoint returns cryptographic keys that are required to, for example,
verify signatures of ID Tokens.

#### Responses

<a id="lists-cryptographic-keys-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description               | Schema                                |
| ------ | -------------------------------------------------------------------------- | ------------------------- | ------------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | jsonWebKeySet             | [jsonWebKeySet](#schemajsonwebkeyset) |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | The standard error format | Inline                                |

<a id="lists-cryptographic-keys-responseschema"></a>

##### Response Schema

Status Code **500**

| Name                        | Type           | Required | Restrictions | Description |
| --------------------------- | -------------- | -------- | ------------ | ----------- |
| » code                      | integer(int64) | false    | none         | none        |
| » details                   | [object]       | false    | none         | none        |
| »» **additionalProperties** | object         | false    | none         | none        |
| » message                   | string         | false    | none         | none        |
| » reason                    | string         | false    | none         | none        |
| » request                   | string         | false    | none         | none        |
| » status                    | string         | false    | none         | none        |

##### Examples

###### 200 response

```json
{
  "keys": [
    {
      "alg": "string",
      "crv": "string",
      "d": "string",
      "dp": "string",
      "dq": "string",
      "e": "string",
      "k": "string",
      "kid": "string",
      "kty": "string",
      "n": "string",
      "p": "string",
      "q": "string",
      "qi": "string",
      "use": "string",
      "x": "string",
      "x5c": ["string"],
      "y": "string"
    }
  ]
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-getWellKnownJSONWebKeys">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-getWellKnownJSONWebKeys-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getWellKnownJSONWebKeys-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getWellKnownJSONWebKeys-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getWellKnownJSONWebKeys-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getWellKnownJSONWebKeys-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getWellKnownJSONWebKeys-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-getWellKnownJSONWebKeys-shell">

```shell
curl -X GET /.well-known/jwks.json \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getWellKnownJSONWebKeys-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/.well-known/jwks.json", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getWellKnownJSONWebKeys-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/.well-known/jwks.json', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getWellKnownJSONWebKeys-java">

```java
// This sample needs improvement.
URL obj = new URL("/.well-known/jwks.json");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getWellKnownJSONWebKeys-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/.well-known/jwks.json',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getWellKnownJSONWebKeys-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/.well-known/jwks.json',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIddecisions"></a>

### Access Control Decision API

```
GET /decisions HTTP/1.1
Accept: application/json

```

> This endpoint works with all HTTP Methods (GET, POST, PUT, ...) and matches
> every path prefixed with /decision.

This endpoint mirrors the proxy capability of ORY Oathkeeper's proxy
functionality but instead of forwarding the request to the upstream server,
returns 200 (request should be allowed), 401 (unauthorized), or 403 (forbidden)
status codes. This endpoint can be used to integrate with other API Proxies like
Ambassador, Kong, Envoy, and many more.

#### Responses

<a id="access-control-decision-api-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description               | Schema |
| ------ | -------------------------------------------------------------------------- | ------------------------- | ------ |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | An empty response         | None   |
| 401    | [Unauthorized](https://tools.ietf.org/html/rfc7235#section-3.1)            | The standard error format | Inline |
| 403    | [Forbidden](https://tools.ietf.org/html/rfc7231#section-6.5.3)             | The standard error format | Inline |
| 404    | [Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)             | The standard error format | Inline |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | The standard error format | Inline |

<a id="access-control-decision-api-responseschema"></a>

##### Response Schema

Status Code **401**

| Name                        | Type           | Required | Restrictions | Description |
| --------------------------- | -------------- | -------- | ------------ | ----------- |
| » code                      | integer(int64) | false    | none         | none        |
| » details                   | [object]       | false    | none         | none        |
| »» **additionalProperties** | object         | false    | none         | none        |
| » message                   | string         | false    | none         | none        |
| » reason                    | string         | false    | none         | none        |
| » request                   | string         | false    | none         | none        |
| » status                    | string         | false    | none         | none        |

Status Code **403**

| Name                        | Type           | Required | Restrictions | Description |
| --------------------------- | -------------- | -------- | ------------ | ----------- |
| » code                      | integer(int64) | false    | none         | none        |
| » details                   | [object]       | false    | none         | none        |
| »» **additionalProperties** | object         | false    | none         | none        |
| » message                   | string         | false    | none         | none        |
| » reason                    | string         | false    | none         | none        |
| » request                   | string         | false    | none         | none        |
| » status                    | string         | false    | none         | none        |

Status Code **404**

| Name                        | Type           | Required | Restrictions | Description |
| --------------------------- | -------------- | -------- | ------------ | ----------- |
| » code                      | integer(int64) | false    | none         | none        |
| » details                   | [object]       | false    | none         | none        |
| »» **additionalProperties** | object         | false    | none         | none        |
| » message                   | string         | false    | none         | none        |
| » reason                    | string         | false    | none         | none        |
| » request                   | string         | false    | none         | none        |
| » status                    | string         | false    | none         | none        |

Status Code **500**

| Name                        | Type           | Required | Restrictions | Description |
| --------------------------- | -------------- | -------- | ------------ | ----------- |
| » code                      | integer(int64) | false    | none         | none        |
| » details                   | [object]       | false    | none         | none        |
| »» **additionalProperties** | object         | false    | none         | none        |
| » message                   | string         | false    | none         | none        |
| » reason                    | string         | false    | none         | none        |
| » request                   | string         | false    | none         | none        |
| » status                    | string         | false    | none         | none        |

##### Examples

###### 401 response

```json
{
  "code": 0,
  "details": [
    {
      "property1": {},
      "property2": {}
    }
  ],
  "message": "string",
  "reason": "string",
  "request": "string",
  "status": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-decisions">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-decisions-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-decisions-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-decisions-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-decisions-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-decisions-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-decisions-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-decisions-shell">

```shell
curl -X GET /decisions \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-decisions-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/decisions", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-decisions-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/decisions', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-decisions-java">

```java
// This sample needs improvement.
URL obj = new URL("/decisions");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-decisions-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/decisions',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-decisions-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/decisions',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdlistRules"></a>

### List all rules

```
GET /rules HTTP/1.1
Accept: application/json

```

This method returns an array of all rules that are stored in the backend. This
is useful if you want to get a full view of what rules you have currently in
place.

<a id="list-all-rules-parameters"></a>

##### Parameters

| Parameter | In    | Type           | Required | Description                             |
| --------- | ----- | -------------- | -------- | --------------------------------------- |
| limit     | query | integer(int64) | false    | The maximum amount of rules returned.   |
| offset    | query | integer(int64) | false    | The offset from where to start looking. |

#### Responses

<a id="list-all-rules-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description               | Schema |
| ------ | -------------------------------------------------------------------------- | ------------------------- | ------ |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | A list of rules           | Inline |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | The standard error format | Inline |

<a id="list-all-rules-responseschema"></a>

##### Response Schema

Status Code **200**

| Name                                                                        | Type                                | Required | Restrictions | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| --------------------------------------------------------------------------- | ----------------------------------- | -------- | ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| _anonymous_                                                                 | [[rule](#schemarule)]               | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| » swaggerRule is a single rule that will get checked on every HTTP request. | [rule](#schemarule)                 | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| »» authenticators                                                           | [[ruleHandler](#schemarulehandler)] | false    | none         | Authenticators is a list of authentication handlers that will try and authenticate the provided credentials. Authenticators are checked iteratively from index 0 to n and if the first authenticator to return a positive result will be the one used. If you want the rule to first check a specific authenticator before "falling back" to others, have that authenticator as the first item in the array.                                                                                                                                                                                                                                                                                              |
| »»» config                                                                  | object                              | false    | none         | Config contains the configuration for the handler. Please read the user guide for a complete list of each handler's available settings.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| »»» handler                                                                 | string                              | false    | none         | Handler identifies the implementation which will be used to handle this specific request. Please read the user guide for a complete list of available handlers.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| »» authorizer                                                               | [ruleHandler](#schemarulehandler)   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| »» description                                                              | string                              | false    | none         | Description is a human readable description of this rule.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| »» id                                                                       | string                              | false    | none         | ID is the unique id of the rule. It can be at most 190 characters long, but the layout of the ID is up to you. You will need this ID later on to update or delete the rule.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| »» match                                                                    | [ruleMatch](#schemarulematch)       | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| »»» methods                                                                 | [string]                            | false    | none         | An array of HTTP methods (e.g. GET, POST, PUT, DELETE, ...). When ORY Oathkeeper searches for rules to decide what to do with an incoming request to the proxy server, it compares the HTTP method of the incoming request with the HTTP methods of each rules. If a match is found, the rule is considered a partial match. If the matchesUrl field is satisfied as well, the rule is considered a full match.                                                                                                                                                                                                                                                                                           |
| »»» url                                                                     | string                              | false    | none         | This field represents the URL pattern this rule matches. When ORY Oathkeeper searches for rules to decide what to do with an incoming request to the proxy server, it compares the full request URL (e.g. https://mydomain.com/api/resource) without query parameters of the incoming request with this field. If a match is found, the rule is considered a partial match. If the matchesMethods field is satisfied as well, the rule is considered a full match. You can use regular expressions in this field to match more than one url. Regular expressions are encapsulated in brackets < and >. The following example matches all paths of the domain `mydomain.com`: `https://mydomain.com/<.*>`. |
| »» mutators                                                                 | [[ruleHandler](#schemarulehandler)] | false    | none         | Mutators is a list of mutation handlers that transform the HTTP request. A common use case is generating a new set of credentials (e.g. JWT) which then will be forwarded to the upstream server. Mutations are performed iteratively from index 0 to n and should all succeed in order for the HTTP request to be forwarded.                                                                                                                                                                                                                                                                                                                                                                             |
| »» upstream                                                                 | [Upstream](#schemaupstream)         | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| »»» preserve_host                                                           | boolean                             | false    | none         | PreserveHost, if false (the default), tells ORY Oathkeeper to set the upstream request's Host header to the hostname of the API's upstream's URL. Setting this flag to true instructs ORY Oathkeeper not to do so.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| »»» strip_path                                                              | string                              | false    | none         | StripPath if set, replaces the provided path prefix when forwarding the requested URL to the upstream URL.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| »»» url                                                                     | string                              | false    | none         | URL is the URL the request will be proxied to.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |

Status Code **500**

| Name                        | Type           | Required | Restrictions | Description |
| --------------------------- | -------------- | -------- | ------------ | ----------- |
| » code                      | integer(int64) | false    | none         | none        |
| » details                   | [object]       | false    | none         | none        |
| »» **additionalProperties** | object         | false    | none         | none        |
| » message                   | string         | false    | none         | none        |
| » reason                    | string         | false    | none         | none        |
| » request                   | string         | false    | none         | none        |
| » status                    | string         | false    | none         | none        |

##### Examples

###### 200 response

```json
[
  {
    "authenticators": [
      {
        "config": {},
        "handler": "string"
      }
    ],
    "authorizer": {
      "config": {},
      "handler": "string"
    },
    "description": "string",
    "id": "string",
    "match": {
      "methods": ["string"],
      "url": "string"
    },
    "mutators": [
      {
        "config": {},
        "handler": "string"
      }
    ],
    "upstream": {
      "preserve_host": true,
      "strip_path": "string",
      "url": "string"
    }
  }
]
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-listRules">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-listRules-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-listRules-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-listRules-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-listRules-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-listRules-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-listRules-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-listRules-shell">

```shell
curl -X GET /rules \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-listRules-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/rules", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-listRules-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/rules', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-listRules-java">

```java
// This sample needs improvement.
URL obj = new URL("/rules");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-listRules-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/rules',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-listRules-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/rules',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdgetRule"></a>

### Retrieve a rule

```
GET /rules/{id} HTTP/1.1
Accept: application/json

```

Use this method to retrieve a rule from the storage. If it does not exist you
will receive a 404 error.

<a id="retrieve-a-rule-parameters"></a>

##### Parameters

| Parameter | In   | Type   | Required | Description |
| --------- | ---- | ------ | -------- | ----------- |
| id        | path | string | true     | none        |

#### Responses

<a id="retrieve-a-rule-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description               | Schema              |
| ------ | -------------------------------------------------------------------------- | ------------------------- | ------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | A rule                    | [rule](#schemarule) |
| 404    | [Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)             | The standard error format | Inline              |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | The standard error format | Inline              |

<a id="retrieve-a-rule-responseschema"></a>

##### Response Schema

Status Code **404**

| Name                        | Type           | Required | Restrictions | Description |
| --------------------------- | -------------- | -------- | ------------ | ----------- |
| » code                      | integer(int64) | false    | none         | none        |
| » details                   | [object]       | false    | none         | none        |
| »» **additionalProperties** | object         | false    | none         | none        |
| » message                   | string         | false    | none         | none        |
| » reason                    | string         | false    | none         | none        |
| » request                   | string         | false    | none         | none        |
| » status                    | string         | false    | none         | none        |

Status Code **500**

| Name                        | Type           | Required | Restrictions | Description |
| --------------------------- | -------------- | -------- | ------------ | ----------- |
| » code                      | integer(int64) | false    | none         | none        |
| » details                   | [object]       | false    | none         | none        |
| »» **additionalProperties** | object         | false    | none         | none        |
| » message                   | string         | false    | none         | none        |
| » reason                    | string         | false    | none         | none        |
| » request                   | string         | false    | none         | none        |
| » status                    | string         | false    | none         | none        |

##### Examples

###### 200 response

```json
{
  "authenticators": [
    {
      "config": {},
      "handler": "string"
    }
  ],
  "authorizer": {
    "config": {},
    "handler": "string"
  },
  "description": "string",
  "id": "string",
  "match": {
    "methods": ["string"],
    "url": "string"
  },
  "mutators": [
    {
      "config": {},
      "handler": "string"
    }
  ],
  "upstream": {
    "preserve_host": true,
    "strip_path": "string",
    "url": "string"
  }
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-getRule">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-getRule-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getRule-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getRule-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getRule-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getRule-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getRule-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-getRule-shell">

```shell
curl -X GET /rules/{id} \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getRule-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/rules/{id}", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getRule-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/rules/{id}', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getRule-java">

```java
// This sample needs improvement.
URL obj = new URL("/rules/{id}");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getRule-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/rules/{id}',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getRule-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/rules/{id}',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="ory-oathkeeper-health"></a>

## health

<a id="opIdisInstanceAlive"></a>

### Check alive status

```
GET /health/alive HTTP/1.1
Accept: application/json

```

This endpoint returns a 200 status code when the HTTP server is up running. This
status does currently not include checks whether the database connection is
working.

If the service supports TLS Edge Termination, this endpoint does not require the
`X-Forwarded-Proto` header to be set.

Be aware that if you are running multiple nodes of this service, the health
status will never refer to the cluster state, only to a single instance.

#### Responses

<a id="check-alive-status-responses"></a>

##### Overview

| Status | Meaning                                                                    | Description               | Schema                              |
| ------ | -------------------------------------------------------------------------- | ------------------------- | ----------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                    | healthStatus              | [healthStatus](#schemahealthstatus) |
| 500    | [Internal Server Error](https://tools.ietf.org/html/rfc7231#section-6.6.1) | The standard error format | Inline                              |

<a id="check-alive-status-responseschema"></a>

##### Response Schema

Status Code **500**

| Name                        | Type           | Required | Restrictions | Description |
| --------------------------- | -------------- | -------- | ------------ | ----------- |
| » code                      | integer(int64) | false    | none         | none        |
| » details                   | [object]       | false    | none         | none        |
| »» **additionalProperties** | object         | false    | none         | none        |
| » message                   | string         | false    | none         | none        |
| » reason                    | string         | false    | none         | none        |
| » request                   | string         | false    | none         | none        |
| » status                    | string         | false    | none         | none        |

##### Examples

###### 200 response

```json
{
  "status": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-isInstanceAlive">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-isInstanceAlive-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceAlive-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceAlive-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceAlive-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceAlive-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceAlive-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-isInstanceAlive-shell">

```shell
curl -X GET /health/alive \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceAlive-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/health/alive", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceAlive-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/health/alive', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceAlive-java">

```java
// This sample needs improvement.
URL obj = new URL("/health/alive");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceAlive-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/health/alive',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceAlive-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/health/alive',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="opIdisInstanceReady"></a>

### Check readiness status

```
GET /health/ready HTTP/1.1
Accept: application/json

```

This endpoint returns a 200 status code when the HTTP server is up running and
the environment dependencies (e.g. the database) are responsive as well.

If the service supports TLS Edge Termination, this endpoint does not require the
`X-Forwarded-Proto` header to be set.

Be aware that if you are running multiple nodes of this service, the health
status will never refer to the cluster state, only to a single instance.

#### Responses

<a id="check-readiness-status-responses"></a>

##### Overview

| Status | Meaning                                                                  | Description          | Schema                                              |
| ------ | ------------------------------------------------------------------------ | -------------------- | --------------------------------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)                  | healthStatus         | [healthStatus](#schemahealthstatus)                 |
| 503    | [Service Unavailable](https://tools.ietf.org/html/rfc7231#section-6.6.4) | healthNotReadyStatus | [healthNotReadyStatus](#schemahealthnotreadystatus) |

##### Examples

###### 200 response

```json
{
  "status": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-isInstanceReady">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-isInstanceReady-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceReady-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceReady-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceReady-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceReady-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-isInstanceReady-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-isInstanceReady-shell">

```shell
curl -X GET /health/ready \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceReady-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/health/ready", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceReady-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/health/ready', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceReady-java">

```java
// This sample needs improvement.
URL obj = new URL("/health/ready");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceReady-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/health/ready',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-isInstanceReady-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/health/ready',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

<a id="ory-oathkeeper-version"></a>

## version

<a id="opIdgetVersion"></a>

### Get service version

```
GET /version HTTP/1.1
Accept: application/json

```

This endpoint returns the service version typically notated using semantic
versioning.

If the service supports TLS Edge Termination, this endpoint does not require the
`X-Forwarded-Proto` header to be set.

Be aware that if you are running multiple nodes of this service, the health
status will never refer to the cluster state, only to a single instance.

#### Responses

<a id="get-service-version-responses"></a>

##### Overview

| Status | Meaning                                                 | Description | Schema                    |
| ------ | ------------------------------------------------------- | ----------- | ------------------------- |
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | version     | [version](#schemaversion) |

##### Examples

###### 200 response

```json
{
  "version": "string"
}
```

<aside class="success">
This operation does not require authentication
</aside>

#### Code samples

<div class="tabs" id="tab-getVersion">
<nav class="tabs-nav">
<ul class="nav nav-tabs au-link-list au-link-list--inline">
<li class="nav-item"><a class="nav-link active" role="tab" href="#tab-getVersion-shell">Shell</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getVersion-go">Go</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getVersion-node">Node.js</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getVersion-java">Java</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getVersion-python">Python</a></li>
<li class="nav-item"><a class="nav-link" role="tab" href="#tab-getVersion-ruby">Ruby</a></li>
</ul>
</nav>
<div class="tab-content">
<div class="tab-pane active" role="tabpanel" id="tab-getVersion-shell">

```shell
curl -X GET /version \
  -H 'Accept: application/json'
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getVersion-go">

```go
package main

import (
    "bytes"
    "net/http"
)

func main() {
    headers := map[string][]string{
        "Accept": []string{"application/json"},
    }

    var body []byte
    // body = ...

    req, err := http.NewRequest("GET", "/version", bytes.NewBuffer(body))
    req.Header = headers

    client := &http.Client{}
    resp, err := client.Do(req)
    // ...
}
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getVersion-node">

```nodejs
const fetch = require('node-fetch');

const headers = {
  'Accept': 'application/json'
}

fetch('/version', {
  method: 'GET',
  headers
})
.then(r => r.json())
.then((body) => {
    console.log(body)
})
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getVersion-java">

```java
// This sample needs improvement.
URL obj = new URL("/version");

HttpURLConnection con = (HttpURLConnection) obj.openConnection();
con.setRequestMethod("GET");

int responseCode = con.getResponseCode();

BufferedReader in = new BufferedReader(
    new InputStreamReader(con.getInputStream())
);

String inputLine;
StringBuffer response = new StringBuffer();
while ((inputLine = in.readLine()) != null) {
    response.append(inputLine);
}
in.close();

System.out.println(response.toString());
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getVersion-python">

```python
import requests

headers = {
  'Accept': 'application/json'
}

r = requests.get(
  '/version',
  params={},
  headers = headers)

print r.json()
```

</div>
<div class="tab-pane" role="tabpanel"  id="tab-getVersion-ruby">

```ruby
require 'rest-client'
require 'json'

headers = {
  'Accept' => 'application/json'
}

result = RestClient.get '/version',
  params: {}, headers: headers

p JSON.parse(result)
```

</div>
</div>
</div>

## Schemas

<a id="tocSupstream">Upstream</a>

#### Upstream

<a id="schemaupstream"></a>

```json
{
  "preserve_host": true,
  "strip_path": "string",
  "url": "string"
}
```

#### Properties

| Name          | Type    | Required | Restrictions | Description                                                                                                                                                                                                        |
| ------------- | ------- | -------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| preserve_host | boolean | false    | none         | PreserveHost, if false (the default), tells ORY Oathkeeper to set the upstream request's Host header to the hostname of the API's upstream's URL. Setting this flag to true instructs ORY Oathkeeper not to do so. |
| strip_path    | string  | false    | none         | StripPath if set, replaces the provided path prefix when forwarding the requested URL to the upstream URL.                                                                                                         |
| url           | string  | false    | none         | URL is the URL the request will be proxied to.                                                                                                                                                                     |

<a id="tocShealthnotreadystatus">healthNotReadyStatus</a>

#### healthNotReadyStatus

<a id="schemahealthnotreadystatus"></a>

```json
{
  "errors": {
    "property1": "string",
    "property2": "string"
  }
}
```

#### Properties

| Name                       | Type   | Required | Restrictions | Description                                                        |
| -------------------------- | ------ | -------- | ------------ | ------------------------------------------------------------------ |
| errors                     | object | false    | none         | Errors contains a list of errors that caused the not ready status. |
| » **additionalProperties** | string | false    | none         | none                                                               |

<a id="tocShealthstatus">healthStatus</a>

#### healthStatus

<a id="schemahealthstatus"></a>

```json
{
  "status": "string"
}
```

#### Properties

| Name   | Type   | Required | Restrictions | Description                  |
| ------ | ------ | -------- | ------------ | ---------------------------- |
| status | string | false    | none         | Status always contains "ok". |

<a id="tocSjsonwebkey">jsonWebKey</a>

#### jsonWebKey

<a id="schemajsonwebkey"></a>

```json
{
  "alg": "string",
  "crv": "string",
  "d": "string",
  "dp": "string",
  "dq": "string",
  "e": "string",
  "k": "string",
  "kid": "string",
  "kty": "string",
  "n": "string",
  "p": "string",
  "q": "string",
  "qi": "string",
  "use": "string",
  "x": "string",
  "x5c": ["string"],
  "y": "string"
}
```

#### Properties

| Name | Type     | Required | Restrictions | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| ---- | -------- | -------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| alg  | string   | false    | none         | The "alg" (algorithm) parameter identifies the algorithm intended for use with the key. The values used should either be registered in the IANA "JSON Web Signature and Encryption Algorithms" registry established by [JWA] or be a value that contains a Collision- Resistant Name.                                                                                                                                                                                                                                                                                              |
| crv  | string   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| d    | string   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| dp   | string   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| dq   | string   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| e    | string   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| k    | string   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| kid  | string   | false    | none         | The "kid" (key ID) parameter is used to match a specific key. This is used, for instance, to choose among a set of keys within a JWK Set during key rollover. The structure of the "kid" value is unspecified. When "kid" values are used within a JWK Set, different keys within the JWK Set SHOULD use distinct "kid" values. (One example in which different keys might use the same "kid" value is if they have different "kty" (key type) values but are considered to be equivalent alternatives by the application using them.) The "kid" value is a case-sensitive string. |
| kty  | string   | false    | none         | The "kty" (key type) parameter identifies the cryptographic algorithm family used with the key, such as "RSA" or "EC". "kty" values should either be registered in the IANA "JSON Web Key Types" registry established by [JWA] or be a value that contains a Collision- Resistant Name. The "kty" value is a case-sensitive string.                                                                                                                                                                                                                                                |
| n    | string   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| p    | string   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| q    | string   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| qi   | string   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| use  | string   | false    | none         | The "use" (public key use) parameter identifies the intended use of the public key. The "use" parameter is employed to indicate whether a public key is used for encrypting data or verifying the signature on data. Values are commonly "sig" (signature) or "enc" (encryption).                                                                                                                                                                                                                                                                                                  |
| x    | string   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| x5c  | [string] | false    | none         | The "x5c" (X.509 certificate chain) parameter contains a chain of one or more PKIX certificates [RFC5280]. The certificate chain is represented as a JSON array of certificate value strings. Each string in the array is a base64-encoded (Section 4 of [RFC4648] -- not base64url-encoded) DER [ITU.X690.1994] PKIX certificate value. The PKIX certificate containing the key value MUST be the first certificate.                                                                                                                                                              |
| y    | string   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |

<a id="tocSjsonwebkeyset">jsonWebKeySet</a>

#### jsonWebKeySet

<a id="schemajsonwebkeyset"></a>

```json
{
  "keys": [
    {
      "alg": "string",
      "crv": "string",
      "d": "string",
      "dp": "string",
      "dq": "string",
      "e": "string",
      "k": "string",
      "kid": "string",
      "kty": "string",
      "n": "string",
      "p": "string",
      "q": "string",
      "qi": "string",
      "use": "string",
      "x": "string",
      "x5c": ["string"],
      "y": "string"
    }
  ]
}
```

#### Properties

| Name | Type                              | Required | Restrictions | Description                                                                                                                                                                                                                                                                           |
| ---- | --------------------------------- | -------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| keys | [[jsonWebKey](#schemajsonwebkey)] | false    | none         | The value of the "keys" parameter is an array of JWK values. By default, the order of the JWK values within the array does not imply an order of preference among them, although applications of JWK Sets can choose to assign a meaning to the order for their purposes, if desired. |

<a id="tocSrule">rule</a>

#### rule

<a id="schemarule"></a>

```json
{
  "authenticators": [
    {
      "config": {},
      "handler": "string"
    }
  ],
  "authorizer": {
    "config": {},
    "handler": "string"
  },
  "description": "string",
  "id": "string",
  "match": {
    "methods": ["string"],
    "url": "string"
  },
  "mutators": [
    {
      "config": {},
      "handler": "string"
    }
  ],
  "upstream": {
    "preserve_host": true,
    "strip_path": "string",
    "url": "string"
  }
}
```

_swaggerRule is a single rule that will get checked on every HTTP request._

#### Properties

| Name           | Type                                | Required | Restrictions | Description                                                                                                                                                                                                                                                                                                                                                                                                  |
| -------------- | ----------------------------------- | -------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| authenticators | [[ruleHandler](#schemarulehandler)] | false    | none         | Authenticators is a list of authentication handlers that will try and authenticate the provided credentials. Authenticators are checked iteratively from index 0 to n and if the first authenticator to return a positive result will be the one used. If you want the rule to first check a specific authenticator before "falling back" to others, have that authenticator as the first item in the array. |
| authorizer     | [ruleHandler](#schemarulehandler)   | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                         |
| description    | string                              | false    | none         | Description is a human readable description of this rule.                                                                                                                                                                                                                                                                                                                                                    |
| id             | string                              | false    | none         | ID is the unique id of the rule. It can be at most 190 characters long, but the layout of the ID is up to you. You will need this ID later on to update or delete the rule.                                                                                                                                                                                                                                  |
| match          | [ruleMatch](#schemarulematch)       | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                         |
| mutators       | [[ruleHandler](#schemarulehandler)] | false    | none         | Mutators is a list of mutation handlers that transform the HTTP request. A common use case is generating a new set of credentials (e.g. JWT) which then will be forwarded to the upstream server. Mutations are performed iteratively from index 0 to n and should all succeed in order for the HTTP request to be forwarded.                                                                                |
| upstream       | [Upstream](#schemaupstream)         | false    | none         | none                                                                                                                                                                                                                                                                                                                                                                                                         |

<a id="tocSrulehandler">ruleHandler</a>

#### ruleHandler

<a id="schemarulehandler"></a>

```json
{
  "config": {},
  "handler": "string"
}
```

#### Properties

| Name    | Type   | Required | Restrictions | Description                                                                                                                                                     |
| ------- | ------ | -------- | ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| config  | object | false    | none         | Config contains the configuration for the handler. Please read the user guide for a complete list of each handler's available settings.                         |
| handler | string | false    | none         | Handler identifies the implementation which will be used to handle this specific request. Please read the user guide for a complete list of available handlers. |

<a id="tocSrulematch">ruleMatch</a>

#### ruleMatch

<a id="schemarulematch"></a>

```json
{
  "methods": ["string"],
  "url": "string"
}
```

#### Properties

| Name    | Type     | Required | Restrictions | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| ------- | -------- | -------- | ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| methods | [string] | false    | none         | An array of HTTP methods (e.g. GET, POST, PUT, DELETE, ...). When ORY Oathkeeper searches for rules to decide what to do with an incoming request to the proxy server, it compares the HTTP method of the incoming request with the HTTP methods of each rules. If a match is found, the rule is considered a partial match. If the matchesUrl field is satisfied as well, the rule is considered a full match.                                                                                                                                                                                                                                                                                           |
| url     | string   | false    | none         | This field represents the URL pattern this rule matches. When ORY Oathkeeper searches for rules to decide what to do with an incoming request to the proxy server, it compares the full request URL (e.g. https://mydomain.com/api/resource) without query parameters of the incoming request with this field. If a match is found, the rule is considered a partial match. If the matchesMethods field is satisfied as well, the rule is considered a full match. You can use regular expressions in this field to match more than one url. Regular expressions are encapsulated in brackets < and >. The following example matches all paths of the domain `mydomain.com`: `https://mydomain.com/<.*>`. |

<a id="tocSversion">version</a>

#### version

<a id="schemaversion"></a>

```json
{
  "version": "string"
}
```

#### Properties

| Name    | Type   | Required | Restrictions | Description                       |
| ------- | ------ | -------- | ------------ | --------------------------------- |
| version | string | false    | none         | Version is the service's version. |
