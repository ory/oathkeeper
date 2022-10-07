// Copyright © 2022 Ory Corp

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
