# OryOathkeeper.RuleApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**createRule**](RuleApi.md#createRule) | **POST** /rules | Create a rule
[**deleteRule**](RuleApi.md#deleteRule) | **DELETE** /rules/{id} | Delete a rule
[**getRule**](RuleApi.md#getRule) | **GET** /rules/{id} | Retrieve a rule
[**listRules**](RuleApi.md#listRules) | **GET** /rules | List all rules
[**updateRule**](RuleApi.md#updateRule) | **PUT** /rules/{id} | Update a rule


<a name="createRule"></a>
# **createRule**
> Rule createRule(opts)

Create a rule

This method allows creation of rules. If a rule id exists, you will receive an error.

### Example
```javascript
var OryOathkeeper = require('ory_oathkeeper');

var apiInstance = new OryOathkeeper.RuleApi();

var opts = { 
  'body': new OryOathkeeper.Rule() // Rule | 
};

var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
};
apiInstance.createRule(opts, callback);
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**Rule**](Rule.md)|  | [optional] 

### Return type

[**Rule**](Rule.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

<a name="deleteRule"></a>
# **deleteRule**
> deleteRule(id)

Delete a rule

Use this endpoint to delete a rule.

### Example
```javascript
var OryOathkeeper = require('ory_oathkeeper');

var apiInstance = new OryOathkeeper.RuleApi();

var id = "id_example"; // String | 


var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully.');
  }
};
apiInstance.deleteRule(id, callback);
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**|  | 

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

Use this method to retrieve a rule from the storage. If it does not exist you will receive a 404 error.

### Example
```javascript
var OryOathkeeper = require('ory_oathkeeper');

var apiInstance = new OryOathkeeper.RuleApi();

var id = "id_example"; // String | 


var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
};
apiInstance.getRule(id, callback);
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**|  | 

### Return type

[**Rule**](Rule.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

<a name="listRules"></a>
# **listRules**
> [Rule] listRules(opts)

List all rules

This method returns an array of all rules that are stored in the backend. This is useful if you want to get a full view of what rules you have currently in place.

### Example
```javascript
var OryOathkeeper = require('ory_oathkeeper');

var apiInstance = new OryOathkeeper.RuleApi();

var opts = { 
  'limit': 789, // Number | The maximum amount of rules returned.
  'offset': 789 // Number | The offset from where to start looking.
};

var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
};
apiInstance.listRules(opts, callback);
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **limit** | **Number**| The maximum amount of rules returned. | [optional] 
 **offset** | **Number**| The offset from where to start looking. | [optional] 

### Return type

[**[Rule]**](Rule.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

<a name="updateRule"></a>
# **updateRule**
> Rule updateRule(id, opts)

Update a rule

Use this method to update a rule. Keep in mind that you need to send the full rule payload as this endpoint does not support patching.

### Example
```javascript
var OryOathkeeper = require('ory_oathkeeper');

var apiInstance = new OryOathkeeper.RuleApi();

var id = "id_example"; // String | 

var opts = { 
  'body': new OryOathkeeper.Rule() // Rule | 
};

var callback = function(error, data, response) {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
};
apiInstance.updateRule(id, opts, callback);
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**|  | 
 **body** | [**Rule**](Rule.md)|  | [optional] 

### Return type

[**Rule**](Rule.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

