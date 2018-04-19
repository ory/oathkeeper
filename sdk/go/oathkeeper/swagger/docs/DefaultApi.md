# \DefaultApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetWellKnown**](DefaultApi.md#GetWellKnown) | **Get** /.well-known/jwks.json | Returns well known keys


# **GetWellKnown**
> JsonWebKeySet GetWellKnown()

Returns well known keys

This endpoint returns public keys for validating the ID tokens issued by ORY Oathkeeper.


### Parameters
This endpoint does not need any parameter.

### Return type

[**JsonWebKeySet**](jsonWebKeySet.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

