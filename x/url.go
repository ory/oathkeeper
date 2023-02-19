// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"net/url"

	"github.com/ory/x/logrusx"
	"github.com/ory/x/urlx"
)

// ParseURLOrPanic parses a url or panics.
// This is the same function as urlx.ParseOrPanic() except that it uses
// urlx.Parse() instead of url.Parse()
func ParseURLOrPanic(in string) *url.URL {
	out, err := urlx.Parse(in)
	if err != nil {
		panic(err.Error())
	}
	return out
}

// ParseURLOrFatal parses a url or fatals.
// This is the same function as urlx.ParseOrFatal() except that it uses
// urlx.Parse() instead of url.Parse()
func ParseURLOrFatal(l *logrusx.Logger, in string) *url.URL {
	out, err := urlx.Parse(in)
	if err != nil {
		l.WithError(err).Fatalf("Unable to parse url: %s", in)
	}
	return out
}
