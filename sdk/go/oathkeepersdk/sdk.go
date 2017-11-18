package oathkeepersdk

import "github.com/ory/oathkeeper/sdk/go/oathkeepersdk/swagger"

type SDK struct {
	*swagger.RuleApi
}

func NewSDK(endpoint string) *SDK {
	return &SDK{
		RuleApi: swagger.NewRuleApiWithBasePath(removeTrailingSlash(endpoint)),
	}
}

func removeTrailingSlash(path string) string {
	for len(path) > 0 && path[len(path)-1] == '/' {
		path = path[0 : len(path)-1]
	}
	return path
}
