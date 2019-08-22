# OryOathkeeper.ApiApi

All URIs are relative to _http://localhost_

| Method                                                           | HTTP request                   | Description                 |
| ---------------------------------------------------------------- | ------------------------------ | --------------------------- |
| [**decisions**](ApiApi.md#decisions)                             | **GET** /decisions             | Access Control Decision API |
| [**getRule**](ApiApi.md#getRule)                                 | **GET** /rules/{id}            | Retrieve a rule             |
| [**getVersion**](ApiApi.md#getVersion)                           | **GET** /version               | Get service version         |
| [**getWellKnownJSONWebKeys**](ApiApi.md#getWellKnownJSONWebKeys) | **GET** /.well-known/jwks.json | Lists cryptographic keys    |
| [**isInstanceAlive**](ApiApi.md#isInstanceAlive)                 | **GET** /health/alive          | Check alive status          |
| [**isInstanceReady**](ApiApi.md#isInstanceReady)                 | **GET** /health/ready          | Check readiness status      |
| [**listRules**](ApiApi.md#listRules)                             | **GET** /rules                 | List all rules              |

<a name="decisions"></a>

# **decisions**

> decisions()

Access Control Decision API

&gt; This endpoint works with all HTTP Methods (GET, POST, PUT, ...) and matches
every path prefixed with /decision. This endpoint mirrors the proxy capability
of ORY Oathkeeper&#39;s proxy functionality but instead of forwarding the
request to the upstream server, returns 200 (request should be allowed), 401
(unauthorized), or 403 (forbidden) status codes. This endpoint can be used to
integrate with other API Proxies like Ambassador, Kong, Envoy, and many more.

### Example

```javascript
var OryOathkeeper = require("ory_oathkeeper");

var apiInstance = new OryOathkeeper.ApiApi();

var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log("API called successfully.");
  }
};
apiInstance.decisions(callback);
```

### Parameters

This endpoint does not need any parameter.

### Return type

null (empty response body)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="getRule"></a>

# **getRule**

> Rule getRule(id)

Retrieve a rule

Use this method to retrieve a rule from the storage. If it does not exist you
will receive a 404 error.

### Example

```javascript
var OryOathkeeper = require("ory_oathkeeper");

var apiInstance = new OryOathkeeper.ApiApi();

var id = "id_example"; // String |

var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log("API called successfully. Returned data: " + data);
  }
};
apiInstance.getRule(id, callback);
```

### Parameters

| Name   | Type       | Description | Notes |
| ------ | ---------- | ----------- | ----- |
| **id** | **String** |             |

### Return type

[**Rule**](Rule.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="getVersion"></a>

# **getVersion**

> Version getVersion()

Get service version

This endpoint returns the service version typically notated using semantic
versioning. If the service supports TLS Edge Termination, this endpoint does not
require the &#x60;X-Forwarded-Proto&#x60; header to be set. Be aware that if you
are running multiple nodes of this service, the health status will never refer
to the cluster state, only to a single instance.

### Example

```javascript
var OryOathkeeper = require("ory_oathkeeper");

var apiInstance = new OryOathkeeper.ApiApi();

var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log("API called successfully. Returned data: " + data);
  }
};
apiInstance.getVersion(callback);
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**Version**](Version.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="getWellKnownJSONWebKeys"></a>

# **getWellKnownJSONWebKeys**

> JsonWebKeySet getWellKnownJSONWebKeys()

Lists cryptographic keys

This endpoint returns cryptographic keys that are required to, for example,
verify signatures of ID Tokens.

### Example

```javascript
var OryOathkeeper = require("ory_oathkeeper");

var apiInstance = new OryOathkeeper.ApiApi();

var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log("API called successfully. Returned data: " + data);
  }
};
apiInstance.getWellKnownJSONWebKeys(callback);
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**JsonWebKeySet**](JsonWebKeySet.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="isInstanceAlive"></a>

# **isInstanceAlive**

> HealthStatus isInstanceAlive()

Check alive status

This endpoint returns a 200 status code when the HTTP server is up running. This
status does currently not include checks whether the database connection is
working. If the service supports TLS Edge Termination, this endpoint does not
require the &#x60;X-Forwarded-Proto&#x60; header to be set. Be aware that if you
are running multiple nodes of this service, the health status will never refer
to the cluster state, only to a single instance.

### Example

```javascript
var OryOathkeeper = require("ory_oathkeeper");

var apiInstance = new OryOathkeeper.ApiApi();

var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log("API called successfully. Returned data: " + data);
  }
};
apiInstance.isInstanceAlive(callback);
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**HealthStatus**](HealthStatus.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="isInstanceReady"></a>

# **isInstanceReady**

> HealthStatus isInstanceReady()

Check readiness status

This endpoint returns a 200 status code when the HTTP server is up running and
the environment dependencies (e.g. the database) are responsive as well. If the
service supports TLS Edge Termination, this endpoint does not require the
&#x60;X-Forwarded-Proto&#x60; header to be set. Be aware that if you are running
multiple nodes of this service, the health status will never refer to the
cluster state, only to a single instance.

### Example

```javascript
var OryOathkeeper = require("ory_oathkeeper");

var apiInstance = new OryOathkeeper.ApiApi();

var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log("API called successfully. Returned data: " + data);
  }
};
apiInstance.isInstanceReady(callback);
```

### Parameters

This endpoint does not need any parameter.

### Return type

[**HealthStatus**](HealthStatus.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

<a name="listRules"></a>

# **listRules**

> [Rule] listRules(opts)

List all rules

This method returns an array of all rules that are stored in the backend. This
is useful if you want to get a full view of what rules you have currently in
place.

### Example

```javascript
var OryOathkeeper = require("ory_oathkeeper");

var apiInstance = new OryOathkeeper.ApiApi();

var opts = {
  limit: 789, // Number | The maximum amount of rules returned.
  offset: 789 // Number | The offset from where to start looking.
};

var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log("API called successfully. Returned data: " + data);
  }
};
apiInstance.listRules(opts, callback);
```

### Parameters

| Name       | Type       | Description                             | Notes      |
| ---------- | ---------- | --------------------------------------- | ---------- |
| **limit**  | **Number** | The maximum amount of rules returned.   | [optional] |
| **offset** | **Number** | The offset from where to start looking. | [optional] |

### Return type

[**[Rule]**](Rule.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json
