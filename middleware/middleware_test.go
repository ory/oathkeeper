// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	grpcTesting "google.golang.org/grpc/test/grpc_testing"
)

func testClient(t *testing.T, l *bufconn.Listener, dialOpts ...grpc.DialOption) grpcTesting.TestServiceClient {
	conn, err := grpc.Dial("bufnet",
		append(dialOpts,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithAuthority("myproject.apis.ory.sh"),
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return l.Dial() }),
		)...,
	)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return grpcTesting.NewTestServiceClient(conn)
}

// testTokenCheckServer is a kratos-style mock server which responds to requests to authenticate the request
// for a successful response the token has to be `Beaerer correct token`.
func testTokenCheckServer(t *testing.T) *httptest.Server {
	s := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("authorization") != "Bearer correct token" {
				t.Logf("denied request %+v", r)
				w.WriteHeader(http.StatusForbidden)
				return
			}
			t.Logf("allowed request %+v", r)
			io.WriteString(w, `{
				"id": "user-id",
				"active": true,
				"expires_at": "2022-12-31T13:50:30.427292Z",
				"authenticated_at": "2022-12-01T13:50:30.825516Z",
				"authenticator_assurance_level": "aal1",
				"authentication_methods": [
				  {
					"method": "password",
					"aal": "aal1",
					"completed_at": "2022-12-01T13:50:30.427375604Z"
				  }
				],
				"issued_at": "2022-12-01T13:50:30.427292Z",
				"identity": {
				  "id": "user-id",
				  "schema_id": "schema-id",
				  "state": "active",
				  "state_changed_at": "2022-12-01T13:50:30.331786Z",
				  "traits": {
					"email": "user@example.org",
					"name": "User"
				  },
				  "verifiable_addresses": [],
				  "recovery_addresses": [],
				  "metadata_public": null,
				  "created_at": "2022-12-01T13:50:30.340643Z",
				  "updated_at": "2022-12-01T13:50:30.340643Z"
				},
				"devices": []
			  }`)
		}))
	t.Cleanup(s.Close)
	return s
}

func writeTestConfig(t *testing.T, pattern string, content string) string {
	f, err := os.CreateTemp(t.TempDir(), pattern)
	if err != nil {
		t.Error(err)
		return ""
	}
	defer f.Close()
	io.WriteString(f, content)

	return f.Name()
}

type testToken string

func (t testToken) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + string(t)}, nil
}
func (t testToken) RequireTransportSecurity() bool { return false }

type upstream struct {
	*MockTestServiceServer
	grpcTesting.UnsafeTestServiceServer
}
