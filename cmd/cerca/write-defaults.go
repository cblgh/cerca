package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gomod.cblgh.org/cerca/database"
	"gomod.cblgh.org/cerca/defaults"
	"gomod.cblgh.org/cerca/util"
)

func writeDefaults() {
	ed := util.Describe("write defaults")

	var confpath string
	var dataDir string
	writeFlags := flag.NewFlagSet("write-defaults", flag.ExitOnError)
	writeFlags.StringVar(&confpath, "config", "", "write cerca's default config to the specified location (including filename). subdirectories are created as needed")
	writeFlags.StringVar(&dataDir, "data-dir", "", "outputs default files to the specified data directory and sets key `data_dir` in config before writing")

	help := createHelpString("write-defaults", []string{
		`cerca write-defaults -config /home/my-user/my-configs/cerca.toml -data-dir /var/lib/cerca`,
	})
	writeFlags.Usage = func() { usage(help, writeFlags) }
	writeFlags.Parse(os.Args[2:])

	// if run without flags, print the help info
	if writeFlags.NFlag() == 0 || len(confpath) == 0 || len(dataDir) == 0 {
		writeFlags.Usage()
		return
	}
	absDataDir, err := filepath.Abs(dataDir)
	ed.Check(err, `could not derive absolute path for --data-dir "%s"`, dataDir)
	dataDir = absDataDir

	keyHash := runAuthKeyGenFunction()
	defaultConfWithAuthKey := strings.Replace(defaults.DEFAULT_CONFIG, `auth_key = ""`, fmt.Sprintf(`auth_key = "%s"`, keyHash), 1)
	defaultConfWithAuthKeyAndData := strings.Replace(defaultConfWithAuthKey, `data_dir = "/var/lib/cerca"`, fmt.Sprintf(`data_dir = "%s"`, dataDir), 1)

	// write the config with auth key and data dir set
	_, err = util.CreateIfNotExist(confpath, defaultConfWithAuthKeyAndData)
	ed.Check(err, "create default config")

	fmt.Printf("wrote config at %s\n", confpath)

	// populate the data dir with initial data
	dbPath := filepath.Join(dataDir, "forum.db")
	docsPath := filepath.Join(dataDir, "docs")
	assetsPath := filepath.Join(dataDir, "assets")

	err = os.MkdirAll(docsPath, 0750)
	ed.Check(err, "could not create '%s'\n", docsPath)

	err = os.MkdirAll(assetsPath, 0750)
	ed.Check(err, "could not create '%s'\n", assetsPath)

	// create the initial database and tables
	_ = database.InitDB(dbPath)
	fmt.Printf("wrote database at %s", dbPath)

	// write the default documents
	type triple struct{ key, docpath, content string }

	triples := []triple{
		{"about", filepath.Join(docsPath, "about.md"), defaults.DEFAULT_ABOUT},
		{"rules", filepath.Join(docsPath, "rules.md"), defaults.DEFAULT_RULES},
		{"registration", filepath.Join(docsPath, "registration.md"), defaults.DEFAULT_REGISTRATION},
		{"logo", filepath.Join(assetsPath, "logo.html"), defaults.DEFAULT_LOGO},
		{"logo.png", filepath.Join(assetsPath, "logo.png"), defaults.DEFAULT_LOGO_PNG},
		{"theme", filepath.Join(assetsPath, "theme.css"), defaults.DEFAULT_THEME},
	}

	for _, t := range triples {
		_, err = util.CreateIfNotExist(t.docpath, t.content)
		ed.Check(err, "could not create %s", t.docpath)
		fmt.Printf("wrote %s\n", t.docpath)
	}
}
