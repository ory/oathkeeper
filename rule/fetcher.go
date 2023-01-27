// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"context"

	"gocloud.dev/blob"
)

type Fetcher interface {
	Watch(ctx context.Context) error
}

type URLMuxSetter interface {
	SetURLMux(mux *blob.URLMux)
}
