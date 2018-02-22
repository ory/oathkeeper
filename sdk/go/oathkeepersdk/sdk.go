// Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
