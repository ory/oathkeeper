// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

import (
	"io"
	"os"
	"testing"
)

// WriteFile writes the content to a new file in a temporary location and
// returns the path. No cleanup is necessary.
func WriteFile(t *testing.T, content string) string {
	f, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Error(err)
		return ""
	}
	defer f.Close()            //nolint:errcheck
	io.WriteString(f, content) //nolint:errcheck,gosec // helper ignores write errors

	return f.Name()
}
