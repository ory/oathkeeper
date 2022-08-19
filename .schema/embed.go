package schema

import "embed"

//go:embed *.json all:pipeline
var FS embed.FS
