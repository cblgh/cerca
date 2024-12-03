package defaults

import (
	_ "embed"
)

//go:embed sample-about.md
var DEFAULT_ABOUT string

//go:embed sample-logo.html
var DEFAULT_LOGO string

//go:embed sample-rules.md
var DEFAULT_RULES string

//go:embed sample-registration-instructions.md
var DEFAULT_REGISTRATION string

//go:embed sample-config.toml
var DEFAULT_CONFIG string
