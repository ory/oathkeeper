// Copyright © 2022 Ory Corp

package spec

import "embed"

//go:embed *.json all:pipeline
var FS embed.FS

//go:embed config.schema.json
var Config []byte
