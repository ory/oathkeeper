// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

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
