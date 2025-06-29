package main

import (
	"github.com/cblgh/cerca/database"
	"flag"
	"fmt"
	"os"
)

func migrate() {
	migrations := map[string]func(string) error{
		"2024-01-password-hash-migration":  database.Migration20240116_PwhashChange,
		"2024-02-thread-private-migration": database.Migration20240720_ThreadPrivateChange,
	}

	var dbPath, migration string
	var listMigrations bool

	migrateFlags := flag.NewFlagSet("migrate", flag.ExitOnError)
	migrateFlags.BoolVar(&listMigrations, "list", false, "list possible migrations")
	migrateFlags.StringVar(&migration, "migration", "", "name of the migration you want to perform on the database")
	migrateFlags.StringVar(&dbPath, "database", "./data/forum.db", "full path to the forum database; e.g. ./data/forum.db")

	help := createHelpString("migrate", []string{
		`cerca migrate -migration 2024-02-thread-private-migration`,
		"cerca migrate -list",
	})
	migrateFlags.Usage = func() { usage(help, migrateFlags) }
	migrateFlags.Parse(os.Args[2:])

	// if run without flags, print the help info
	if migrateFlags.NFlag() == 0 {
		migrateFlags.Usage()
		return
	}

	if listMigrations {
		inform("Possible migrations:")
		for key := range migrations {
			fmt.Println("\t", key)
		}
		os.Exit(0)
	}

	if migration == "" {
		complain(help)
	} else if _, ok := migrations[migration]; !ok {
		complain(fmt.Sprintf("chosen migration »%s» does not match one of the available migrations. see migrations with flag --list", migration))
	}

	// check if database exists! we dont wanna create a new db in this case ':)
	if !database.CheckExists(dbPath) {
		complain("couldn't find database at %s", dbPath)
	}

	// perform migration
	err := migrations[migration](dbPath)
	if err == nil {
		inform(fmt.Sprintf("Migration »%s» completed", migration))
	} else {
		complain("migration terminated early due to error")
	}
}
