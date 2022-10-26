package types

import (
	"path/filepath"
)

type Config struct {
	Community struct {
		Name        string `json:"name"`
		ConductLink string `json:"conduct_url"` // optional
		Language    string `json:"language"`
	} `json:"general"`

	Documents struct {
		LogoPath                    string `json:"logo"`
		AboutPath                   string `json:"about"`
		RegisterRulesPath           string `json:"rules"`
		VerificationExplanationPath string `json:"verification_instructions"`
	} `json:"documents"`
}

// Ensure that, at the very least, default paths exist for each expected document path. Does not overwrite previously set values.
func (c *Config) EnsureDefaultPaths() {
	if c.Documents.AboutPath == "" {
		c.Documents.AboutPath = filepath.Join("content", "about.md")
	}
	if c.Documents.RegisterRulesPath == "" {
		c.Documents.RegisterRulesPath = filepath.Join("content", "rules.md")
	}
	if c.Documents.VerificationExplanationPath == "" {
		c.Documents.VerificationExplanationPath = filepath.Join("content", "verification-instructions.md")
	}
	if c.Documents.LogoPath == "" {
		c.Documents.LogoPath = filepath.Join("content", "logo.html")
	}
}

/*
config structure:
["general"]
name = "Merveilles"
conduct_link = "https://github.com/merveilles/Resources/blob/master/CONDUCT.md"
language = "English"


["documents"]
logo = "./logo.svg"
about = "./about.md"
rules = "./rules.md"
verification_instructions = "./verification-instructions.md"
*/
