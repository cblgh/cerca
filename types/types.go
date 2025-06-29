package types

import (
	"path/filepath"
	"log"
)

type Config struct {
	Tooling struct {
		CercaRoot   string `json:"cerca_root"`
	} `json:"application"`

	Community struct {
		Name        string `json:"name"`
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
		c.Documents.AboutPath = c.JoinWithRoot("content", "about.md")
	}
	if c.Documents.RegisterRulesPath == "" {
		c.Documents.RegisterRulesPath = c.JoinWithRoot("content", "rules.md")
	}
	if c.Documents.RegistrationExplanationPath == "" {
		c.Documents.RegistrationExplanationPath = c.JoinWithRoot("content", "registration-instructions.md")
	}
	if c.Documents.LogoPath == "" {
		c.Documents.LogoPath = c.JoinWithRoot("content", "logo.html")
	}
}


// JoinWithRoot takes into account whether `path` is absolute or not. If it is absolute, then just return the absolute
// path joined as a single string instead.
func (c *Config) JoinWithRoot(path ...string) string {
	appRoot := c.Tooling.CercaRoot
	if appRoot == "" {
		appRoot = "./"
	}
	// inlined util.JoinWithBase because of import cycle
	joinedPath := filepath.Join(path...)
	if filepath.IsAbs(joinedPath) {
		return joinedPath
	}
	var p []string
	p = append(p, appRoot)
	p = append(p, joinedPath)
	finalPath, err := filepath.Abs(filepath.Join(p...))
	if err != nil {
		log.Fatalf("types.go (JoinWithRoot): %w\n", err)
	}
	return finalPath
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
