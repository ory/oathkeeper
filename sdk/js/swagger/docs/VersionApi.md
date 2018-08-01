# OryOathkeeper.VersionApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getVersion**](VersionApi.md#getVersion) | **GET** /version | Get the version of Oathkeeper


<a name="getVersion"></a>
# **getVersion**
> Version getVersion()

Get the version of Oathkeeper

This endpoint returns the version as &#x60;{ \&quot;version\&quot;: \&quot;VERSION\&quot; }&#x60;. The version is only correct with the prebuilt binary and not custom builds.

### Example
```javascript
var OryOathkeeper = require('ory_oathkeeper');

var apiInstance = new OryOathkeeper.VersionApi();

var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
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

