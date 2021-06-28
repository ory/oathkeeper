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

package rule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindStringSubmatch(t *testing.T) {
	type args struct {
		pattern      string
		matchAgainst string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "bad pattern",
			args: args{
				pattern:      `urn:foo:<.?>`,
				matchAgainst: "urn:foo:user",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "one group",
			args: args{
				pattern:      `urn:foo:<.*>`,
				matchAgainst: "urn:foo:user",
			},
			want:    []string{"user"},
			wantErr: false,
		},
		{
			name: "several groups",
			args: args{
				pattern:      `urn:foo:<.*>:<.*>`,
				matchAgainst: "urn:foo:user:one",
			},
			want:    []string{"user", "one"},
			wantErr: false,
		},
		{
			name: "classic foo bar",
			args: args{
				pattern:      `urn:foo:<foo|bar>`,
				matchAgainst: "urn:foo:bar",
			},
			want:    []string{"bar"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regexpEngine := new(regexpMatchingEngine)
			got, err := regexpEngine.FindStringSubmatch(tt.args.pattern, tt.args.matchAgainst)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindStringSubmatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.ElementsMatch(t, got, tt.want, "FindStringSubmatch() got = %v, want %v", got, tt.want)
		})
	}
}
