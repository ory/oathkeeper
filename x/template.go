// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
