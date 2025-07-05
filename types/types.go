package types

import (
	"path/filepath"
	"fmt"
	"os"
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
}

// Ensure that, at the very least, default paths exist for each expected document path. 
func (c *Config) EnsureDefaultPaths() {
	docsPath := filepath.Join(c.General.DataDir, "docs")
	assetsPath := filepath.Join(c.General.DataDir, "assets")
	err := os.MkdirAll(docsPath, 0750)
	if err != nil {
		fmt.Printf("could not create '%s'\n", docsPath)
	}
	err = os.MkdirAll(assetsPath, 0750)
	if err != nil {
		fmt.Printf("could not create '%s'\n", assetsPath)
	}
}

/*
config structure:
[general]
name = "Merveilles"
conduct_link = "https://github.com/merveilles/Resources/blob/master/CONDUCT.md"
language = "English"

[rss]
feed_name = "Merveilles Forum"
feed_description = "marvellous happenings and introspective wanderings"
forum_url = "https://forum.merveilles.town"

*/
