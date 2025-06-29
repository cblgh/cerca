package types

import (
	"path/filepath"
)

type Config struct {
	General struct {
		Name        string `json:"name"`
		DataDir     string `json:"data_dir"`
		AuthKey     string `json:"auth_key"`
		ConductLink string `json:"conduct_url"` // optional
		Language    string `json:"language"`
	} `json:"general"`

	RSS struct {
		Name        string `json:"feed_name"`
		Description string `json:"feed_description"`
		URL         string `json:"forum_url"`
	} `json:"rss"`

	Documents struct {
		LogoPath                    string `json:"logo"`
		AboutPath                   string `json:"about"`
		RegisterRulesPath           string `json:"rules"`
		RegistrationExplanationPath string `json:"registration_instructions"`
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
	if c.Documents.RegistrationExplanationPath == "" {
		c.Documents.RegistrationExplanationPath = filepath.Join("content", "registration-instructions.md")
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
registration_instructions = "./registration-instructions.md"
*/
