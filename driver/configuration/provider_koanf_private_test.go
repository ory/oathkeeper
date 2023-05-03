// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package configuration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/logrusx"
)

func TestGetURL(t *testing.T) {
	kp, err := NewKoanfProvider(context.Background(), nil, logrusx.New("", ""))
	require.NoError(t, err)
	assert.Nil(t, kp.getURL("", "key"))
	assert.Nil(t, kp.getURL("a", "key"))
}
