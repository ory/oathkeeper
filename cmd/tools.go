// +build tools

package cmd

import (
	_ "github.com/go-swagger/go-swagger/cmd/swagger"
	_ "github.com/mattn/goveralls"
	_ "github.com/sqs/goreturns"
	_ "golang.org/x/tools/cmd/cover"

	_ "github.com/ory/go-acc"
)
