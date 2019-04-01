// +build tools

package cmd

import (
	_ "github.com/mattn/goveralls"
	_ "github.com/mitchellh/gox"
	_ "github.com/ory/go-acc"
	_ "github.com/tcnksm/ghr"
	_ "golang.org/x/tools/cmd/cover"
	_ "github.com/go-swagger/go-swagger/cmd/swagger"
	_ "golang.org/x/tools/cmd/goimports"
)
