// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

//nolint:govet // keep legacy build tag for older Go toolchains
//go:build tools
// +build tools

package main

import (
	_ "github.com/go-swagger/go-swagger/cmd/swagger"

	_ "github.com/ory/go-acc"
)
