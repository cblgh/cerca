package main

import (
	"cerca/constants"
	"cerca/database"
	"flag"
	"fmt"
	"os"
)

func admin() {
	var username string
	var forumDomain string
	var dbPath string

	adminCmd := flag.NewFlagSet("admin", flag.ExitOnError)
	adminCmd.StringVar(&forumDomain, "url", "https://forum.merveilles.town", "root url to forum, referenced in output")
	adminCmd.StringVar(&username, "username", "", "username who should be made admin")
	adminCmd.StringVar(&dbPath, "database", "./data/forum.db", "full path to the forum database; e.g. ./data/forum.db")

	help := createHelpString([]string{
		"cerca admin -username myCoolUsername",
	}, false)
	adminCmd.Usage = func() { usage(help, adminCmd) }
	adminCmd.Parse(os.Args[2:])

	adminRoute := fmt.Sprintf("%s/admin", forumDomain)

	if username == "" {
		complain(help)
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
