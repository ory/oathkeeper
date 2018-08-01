# OryOathkeeper.DefaultApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getWellKnown**](DefaultApi.md#getWellKnown) | **GET** /.well-known/jwks.json | Returns well known keys


<a name="getWellKnown"></a>
# **getWellKnown**
> JsonWebKeySet getWellKnown()

Returns well known keys

This endpoint returns public keys for validating the ID tokens issued by ORY Oathkeeper.

### Example
```javascript
var OryOathkeeper = require('ory_oathkeeper');

var apiInstance = new OryOathkeeper.DefaultApi();

var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
};
apiInstance.getWellKnown(callback);
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

