// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline/authn"
	. "github.com/ory/oathkeeper/pipeline/authz"
	"github.com/ory/oathkeeper/rule"
	"github.com/ory/x/configx"
	"github.com/ory/x/logrusx"
)

func TestAuthorizerAllowIfAuthorize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		config  json.RawMessage
		session *authn.AuthenticationSession
		wantErr bool
	}{
		{
			name:   "ok",
			config: json.RawMessage(`{"key":"{{ .Subject }}","op":"equals","value":"alice"}`),
			session: &authn.AuthenticationSession{
				Subject: "alice",
			},
			wantErr: false,
		},
		{
			name:   "invalid operator",
			config: json.RawMessage(`{"key":"{{ .Subject }}","op":"strange-operator","value":"alice"}`),
			session: &authn.AuthenticationSession{
				Subject: "alice",
			},
			wantErr: true,
		},
		{
			name:   "forbidden",
			config: json.RawMessage(`{"key":"{{ .Subject }}","op":"equals","value":"alice"}`),
			session: &authn.AuthenticationSession{
				Subject: "bob",
			},
			wantErr: true,
		},
		{
			name:   "invalid comparison type",
			config: json.RawMessage(`{"key":"{{ .Extra.groups }}","op":"equals","value":["alice", "bob"]}`),
			session: &authn.AuthenticationSession{
				Extra: map[string]interface{}{
					"groups": []string{"alice", "bob"},
				},
			},
			wantErr: true,
		},
		{
			name:   "ok with string slice as string",
			config: json.RawMessage(`{"key":"{{ .Extra.groups }}","op":"equals","value":"[alice bob]"}`),
			session: &authn.AuthenticationSession{
				Extra: map[string]interface{}{
					"groups": []string{"alice", "bob"},
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := configuration.NewKoanfProvider(
				context.Background(), nil, logrusx.New("", ""),
				configx.SkipValidation())
			require.NoError(t, err)
			a := NewAuthorizerAllowIf(p)
			p.SetForTest(t, configuration.AuthorizerAllowIfIsEnabled, true)

			if err := a.Authorize(&http.Request{}, tc.session, tc.config, &rule.Rule{}); (err != nil) != tc.wantErr {
				t.Errorf("Authorize() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestAuthorizerAllowIfValidatet(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		config  json.RawMessage
		wantErr bool
	}{
		{
			name:    "valid config with operator equals",
			enabled: true,
			config:  json.RawMessage(`{"key":"my-key","value":"my-value","op":"equals"}`),
			wantErr: false,
		},
		{
			name:    "disabled",
			enabled: false,
			config:  json.RawMessage(`{}`),
			wantErr: true,
		},
		{
			name:    "empty configuration",
			enabled: true,
			config:  json.RawMessage(`{}`),
			wantErr: true,
		},
		{
			name:    "missing key",
			enabled: true,
			config:  json.RawMessage(`{"value":"my-value","operator":"equals"}`),
			wantErr: true,
		},
		{
			name:    "missing value",
			enabled: true,
			config:  json.RawMessage(`{"key":"my-key","operator":"equals"}`),
			wantErr: true,
		},
		{
			name:    "missing operator",
			enabled: true,
			config:  json.RawMessage(`{"key":"my-key","value":"my-value"}`),
			wantErr: true,
		},
		{
			name:    "invalid operator",
			enabled: true,
			config:  json.RawMessage(`{"key":"my-key","value":"my-value","op":"invalid"}`),
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := configuration.NewKoanfProvider(
				context.Background(), nil, logrusx.New("", ""),
				configx.SkipValidation())
			require.NoError(t, err)
			a := NewAuthorizerAllowIf(p)
			p.SetForTest(t, configuration.AuthorizerAllowIfIsEnabled, tc.enabled)
			if err := a.Validate(tc.config); (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
