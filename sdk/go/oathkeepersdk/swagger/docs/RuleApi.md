# \RuleApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**CreateRule**](RuleApi.md#CreateRule) | **Post** /rules | Create a rule
[**DeleteRule**](RuleApi.md#DeleteRule) | **Delete** /rules/{id} | Delete a rule
[**GetRule**](RuleApi.md#GetRule) | **Get** /rules/{id} | Retrieve a rule
[**ListRules**](RuleApi.md#ListRules) | **Get** /rules | List all rules
[**UpdateRule**](RuleApi.md#UpdateRule) | **Put** /rules/{id} | Update a rule


# **CreateRule**
> Rule CreateRule($body)

Create a rule

This method allows creation of rules. If a rule id exists, you will receive an error.


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**Rule**](Rule.md)|  | [optional] 

### Return type

[**Rule**](rule.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteRule**
> DeleteRule($id)

Delete a rule

Use this endpoint to delete a rule.


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string**|  | 

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetRule**
> Rule GetRule($id)

Retrieve a rule

Use this method to retrieve a rule from the storage. If it does not exist you will receive a 404 error.


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string**|  | 

### Return type

[**Rule**](rule.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **ListRules**
> []Rule ListRules()

List all rules

This method returns an array of all rules that are stored in the backend. This is useful if you want to get a full view of what rules you have currently in place.


### Parameters
This endpoint does not need any parameter.

### Return type

[**[]Rule**](rule.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **UpdateRule**
> Rule UpdateRule($id, $body)

Update a rule

Use this method to update a rule. Keep in mind that you need to send the full rule payload as this endpoint does not support patching.


### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **string**|  | 
 **body** | [**Rule**](Rule.md)|  | [optional] 

### Return type

[**Rule**](rule.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

