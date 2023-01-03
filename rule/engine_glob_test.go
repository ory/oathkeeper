// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelimiters(t *testing.T) {
	var tests = []struct {
		input string
		out   []int
		err   error
	}{
		{
			input: "<",
			err:   ErrUnbalancedPattern,
		},
		{
			input: ">",
			err:   ErrUnbalancedPattern,
		},
		{
			input: ">>",
			err:   ErrUnbalancedPattern,
		},
		{
			input: "><>",
			err:   ErrUnbalancedPattern,
		},
		{
			input: "foo.bar<var",
			err:   ErrUnbalancedPattern,
		},
		{
			input: "foo.bar>var",
			err:   ErrUnbalancedPattern,
		},
		{
			input: "foo.bar><var",
			err:   ErrUnbalancedPattern,
		},
		{
			input: "foo.bar<<>var",
			err:   ErrUnbalancedPattern,
		},
		{
			input: "foo.bar<<>>",
			out: []int{
				7, 11,
			},
		},
		{
			input: "foo.bar<<>><>",
			out: []int{
				7, 11,
				11, 13,
			},
		},
		{
			input: "foo.bar<<>><>tt<>",
			out: []int{
				7, 11,
				11, 13,
				15, 17,
			},
		},
	}

	for tn, tc := range tests {
		t.Run(strconv.Itoa(tn), func(t *testing.T) {
			out, err := delimiterIndices(tc.input, '<', '>')
			assert.Equal(t, tc.out, out)
			assert.Equal(t, tc.err, err)

		})
	}
}

func TestIsMatch(t *testing.T) {
	type args struct {
		pattern      string
		matchAgainst string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "question mark1",
			args: args{
				pattern:      `urn:foo:<?>`,
				matchAgainst: "urn:foo:user",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "question mark2",
			args: args{
				pattern:      `urn:foo:<?>`,
				matchAgainst: "urn:foo:u",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "question mark3",
			args: args{
				pattern:      `urn:foo:<?>`,
				matchAgainst: "urn:foo:",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "question mark4",
			args: args{
				pattern:      `urn:foo:<?>&&<?>`,
				matchAgainst: "urn:foo:w&&r",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "question mark5 - both as a special char and a literal",
			args: args{
				pattern:      `urn:foo:<?>?<?>`,
				matchAgainst: "urn:foo:w&r",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "question mark5 - both as a special char and a literal1",
			args: args{
				pattern:      `urn:foo:<?>?<?>`,
				matchAgainst: "urn:foo:w?r",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "asterisk",
			args: args{
				pattern:      `urn:foo:<*>`,
				matchAgainst: "urn:foo:user",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "asterisk1",
			args: args{
				pattern:      `urn:foo:<*>`,
				matchAgainst: "urn:foo:",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "asterisk2",
			args: args{
				pattern:      `urn:foo:<*>:<*>`,
				matchAgainst: "urn:foo:usr:swen",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "asterisk: both as a special char and a literal",
			args: args{
				pattern:      `*:foo:<*>:<*>`,
				matchAgainst: "urn:foo:usr:swen",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "asterisk: both as a special char and a literal1",
			args: args{
				pattern:      `*:foo:<*>:<*>`,
				matchAgainst: "*:foo:usr:swen",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "asterisk + question mark",
			args: args{
				pattern:      `urn:foo:<*>:role:<?>`,
				matchAgainst: "urn:foo:usr:role:a",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "asterisk + question mark1",
			args: args{
				pattern:      `urn:foo:<*>:role:<?>`,
				matchAgainst: "urn:foo:usr:role:admin",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "square brackets",
			args: args{
				pattern:      `urn:foo:<m[a,o,u]n>`,
				matchAgainst: "urn:foo:moon",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "square brackets1",
			args: args{
				pattern:      `urn:foo:<m[a,o,u]n>`,
				matchAgainst: "urn:foo:man",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "square brackets2",
			args: args{
				pattern:      `urn:foo:<m[!a,o,u]n>`,
				matchAgainst: "urn:foo:man",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "square brackets3",
			args: args{
				pattern:      `urn:foo:<m[!a,o,u]n>`,
				matchAgainst: "urn:foo:min",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "asterisk matches only one path segment",
			args: args{
				pattern:      `http://example.com/<*>`,
				matchAgainst: "http://example.com/foo/bar",
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globEngine := new(globMatchingEngine)
			got, err := globEngine.IsMatching(tt.args.pattern, tt.args.matchAgainst)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsMatching() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsMatching() got = %v, want %v", got, tt.want)
			}
		})
	}
}
