# OryOathkeeper.JudgeApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**judge**](JudgeApi.md#judge) | **GET** /judge | Judge if a request should be allowed or not


<a name="judge"></a>
# **judge**
> judge()

Judge if a request should be allowed or not

This endpoint mirrors the proxy capability of ORY Oathkeeper&#39;s proxy functionality but instead of forwarding the request to the upstream server, returns 200 (request should be allowed), 401 (unauthorized), or 403 (forbidden) status codes. This endpoint can be used to integrate with other API Proxies like Ambassador, Kong, Envoy, and many more.

### Example
```javascript
var OryOathkeeper = require('ory_oathkeeper');

var apiInstance = new OryOathkeeper.JudgeApi();

var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully.');
  }
};
apiInstance.judge(callback);
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

