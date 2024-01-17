package main

import (
	"cerca/database"
	"flag"
	"fmt"
	"os"
)

func inform(msg string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf("%s\n", fmt.Sprintf(msg, args...))
	} else {
		fmt.Printf("%s\n", msg)
	}
}

func complain(msg string, args ...interface{}) {
	if len(args) > 0 {
		inform(msg, args)
	} else {
		inform(msg)
	}
	os.Exit(0)
}


func main() {
	migrations := map[string]func(string) error {"2024-01-password-hash-migration": database.Migration20240116_PwhashChange}

	var dbPath, migration string
	var listMigrations bool
	flag.BoolVar(&listMigrations, "list", false, "list possible migrations")
	flag.StringVar(&migration, "migration", "", "name of the migration you want to perform on the database")
	flag.StringVar(&dbPath, "database", "./data/forum.db", "full path to the forum database; e.g. ./data/forum.db")
	flag.Parse()

	usage := `usage
	migration-tool --list 
	migration-tool --migration <name of migration>
  `

	if listMigrations {
		inform("Possible migrations:")
		for key := range migrations {
			fmt.Println("\t", key)
		}
		os.Exit(0)
	}

	if migration == "" {
		complain(usage)
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
