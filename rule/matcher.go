// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"context"
	"net/url"
)

type (
	Protocol string

	Matcher interface {
		Match(ctx context.Context, method string, u *url.URL, protocol Protocol) (*Rule, error)
	}
)

const (
	ProtocolHTTP Protocol = "http"
	ProtocolGRPC Protocol = "grpc"
)
