# Rule

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Authenticators** | [**[]RuleHandler**](ruleHandler.md) | Authenticators is a list of authentication handlers that will try and authenticate the provided credentials. Authenticators are checked iteratively from index 0 to n and if the first authenticator to return a positive result will be the one used.  If you want the rule to first check a specific authenticator  before \&quot;falling back\&quot; to others, have that authenticator as the first item in the array. | [optional] [default to null]
**Authorizer** | [**RuleHandler**](ruleHandler.md) |  | [optional] [default to null]
**CredentialsIssuer** | [**RuleHandler**](ruleHandler.md) |  | [optional] [default to null]
**Description** | **string** | Description is a human readable description of this rule. | [optional] [default to null]
**Id** | **string** | ID is the unique id of the rule. It can be at most 190 characters long, but the layout of the ID is up to you. You will need this ID later on to update or delete the rule. | [optional] [default to null]
**Match** | [**RuleMatch**](ruleMatch.md) |  | [optional] [default to null]
**Upstream** | [**Upstream**](Upstream.md) |  | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


