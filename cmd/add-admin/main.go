package main

import (
	"cerca/database"
	"cerca/constants"
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
	var username string
	var forumDomain string
	var dbPath string
	flag.StringVar(&forumDomain, "url", "https://forum.merveilles.town", "root url to forum, referenced in output")
	flag.StringVar(&username, "username", "", "username who should be made admin")
	flag.StringVar(&dbPath, "database", "./data/forum.db", "full path to the forum database; e.g. ./data/forum.db")
	flag.Parse()

	usage := `usage
	add-admin --username <username to make admin> --url <rool url to forum> --database ./data/forum.db
	add-admin --help for more information
  `

	adminRoute := fmt.Sprintf("%s/admin", forumDomain)

	if username == "" {
		complain(usage)
	}

	// check if database exists! we dont wanna create a new db in this case ':)
	if !database.CheckExists(dbPath) {
		complain("couldn't find database at %s", dbPath)
	}

	db := database.InitDB(dbPath)

	userid, err := db.GetUserID(username)
	if err != nil {
		complain("username %s not in database", username)
	}
	inform("Attempting to make %s (id %d) admin...", username, userid)
	err = db.AddAdmin(userid)
	if err != nil {
		complain("Something went wrong: %s", err)
	}

	// log cmd actions just as admin web-actions are logged
	systemUserid := db.GetSystemUserid()
	err = db.AddModerationLog(systemUserid, userid, constants.MODLOG_ADMIN_MAKE)
	if err != nil {
		complain("adding mod log for adding new admin failed (%w)", err)
	}

	inform("Successfully added %s (id %d) as an admin", username, userid)
	inform("Please visit %s for all your administration needs (changing usernames, resetting passwords, deleting user accounts)", adminRoute)
	inform("Admin action has been logged to /moderations")
}
