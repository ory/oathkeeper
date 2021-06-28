// Copyright 2021 Ory GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package x

import (
	"fmt"
	"reflect"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// NewTemplate creates a template with additional functions
func NewTemplate(id string) *template.Template {
	return template.New(id).
		// Implies that zero value will be used if a key is missing.
		Option("missingkey=zero").
		Funcs(template.FuncMap{
			"print": func(i interface{}) string {
				if i == nil {
					return ""
				}
				return fmt.Sprintf("%v", i)
			},
			"printIndex": func(element interface{}, i int) string {
				if element == nil {
					return ""
				}

				list := reflect.ValueOf(element)

				if list.Kind() == reflect.Slice && i < list.Len() {
					return fmt.Sprintf("%v", list.Index(i))
				}

				return ""
			},
		}).
		Funcs(sprig.TxtFuncMap())
}
