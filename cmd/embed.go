package cmd

import (
	"embed"
)

//go:embed *
var embeddedFS embed.FS
