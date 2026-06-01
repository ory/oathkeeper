// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package credentials

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScope(t *testing.T) {
	for k, tc := range []struct {
		i  map[string]interface{}
		ev []string
		ek string
	}{
		{i: map[string]interface{}{}, ev: []string{}},
		{i: map[string]interface{}{"scp": "foo bar"}, ev: []string{"foo", "bar"}, ek: "scp"},
		{i: map[string]interface{}{"scope": "foo bar"}, ev: []string{"foo", "bar"}, ek: "scope"},
		{i: map[string]interface{}{"scopes": "foo bar"}, ev: []string{"foo", "bar"}, ek: "scopes"},
		{i: map[string]interface{}{"scp": []string{"foo", "bar"}}, ev: []string{"foo", "bar"}, ek: "scp"},
		{i: map[string]interface{}{"scope": []string{"foo", "bar"}}, ev: []string{"foo", "bar"}, ek: "scope"},
		{i: map[string]interface{}{"scopes": []string{"foo", "bar"}}, ev: []string{"foo", "bar"}, ek: "scopes"},
		{i: map[string]interface{}{"scp": []interface{}{"foo", "bar"}}, ev: []string{"foo", "bar"}, ek: "scp"},
		{i: map[string]interface{}{"scope": []interface{}{"foo", "bar"}}, ev: []string{"foo", "bar"}, ek: "scope"},
		{i: map[string]interface{}{"scopes": []interface{}{"foo", "bar"}}, ev: []string{"foo", "bar"}, ek: "scopes"},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			value, key := scope(tc.i)
			assert.EqualValues(t, tc.ev, value)
			assert.EqualValues(t, tc.ek, key)
		})
	}
}
