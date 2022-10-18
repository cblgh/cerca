package defaults

import (
	_ "embed"
)

//go:embed sample-about.md
var DEFAULT_ABOUT string

//go:embed sample-logo.svg
var DEFAULT_LOGO string

//go:embed sample-rules.md
var DEFAULT_RULES string

//go:embed sample-verification-instructions.md
var DEFAULT_VERIFICATION string

//go:embed sample-config.toml
var DEFAULT_CONFIG string
