// Copyright © 2022 Ory Corp

//go:build tools
// +build tools

package main

import (
	_ "github.com/go-swagger/go-swagger/cmd/swagger"
	_ "github.com/mattn/goveralls"
	_ "github.com/sqs/goreturns"
	_ "golang.org/x/tools/cmd/cover"

	_ "github.com/gorilla/websocket"

	_ "github.com/ory/go-acc"

	_ "github.com/ory/cli"
)
