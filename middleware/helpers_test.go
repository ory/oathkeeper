// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeTestConfigf(t *testing.T, format string, args ...any) string {
	f, err := os.Create(filepath.Join(t.TempDir(), "file.yaml"))
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	_, _ = fmt.Fprintf(f, format, args...)

	return f.Name()
}
