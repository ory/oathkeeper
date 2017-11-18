# Rule

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowAnonymousModeEnabled** | **bool** | AllowAnonymousModeEnabled sets if the endpoint is public, thus not needing any authorization at all. | [optional] [default to null]
**BasicAuthorizationModeEnabled** | **bool** | BasicAuthorizationModeEnabled if set true disables checking access control policies. | [optional] [default to null]
**Description** | **string** | Description describes the rule. | [optional] [default to null]
**Id** | **string** | ID the a unique id of a rule. | [optional] [default to null]
**MatchesMethods** | **[]string** | MatchesMethods is a list of HTTP methods that this rule matches. | [optional] [default to null]
**MatchesUrl** | **string** | MatchesURL is a regular expression of paths this rule matches. | [optional] [default to null]
**PassThroughModeEnabled** | **bool** | PassThroughModeEnabled if set true disables firewall capabilities. | [optional] [default to null]
**RequiredAction** | **string** | RequiredScopes is the action this rule requires. | [optional] [default to null]
**RequiredResource** | **string** | RequiredScopes is the resource this rule requires. | [optional] [default to null]
**RequiredScopes** | **[]string** | RequiredScopes is a list of scopes that are required by this rule. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


