package main

import (
	"github.com/cblgh/cerca/constants"
	"github.com/cblgh/cerca/database"
	"flag"
	"fmt"
	"os"
)

func admin() {
	var username string
	var forumDomain string
	var dbPath string

	adminFlags := flag.NewFlagSet("makeadmin", flag.ExitOnError)
	adminFlags.StringVar(&forumDomain, "url", "https://forum.merveilles.town", "root url to forum, referenced in output")
	adminFlags.StringVar(&username, "username", "", "username who should be made admin")
	adminFlags.StringVar(&dbPath, "database", "./data/forum.db", "full path to the forum database; e.g. ./data/forum.db")

	help := createHelpString("makeadmin", []string{
		`cerca makeadmin -username "<existing username>"`,
	})
	adminFlags.Usage = func() { usage(help, adminFlags) }
	adminFlags.Parse(os.Args[2:])

	// if run without flags, print the help info
	if adminFlags.NFlag() == 0 {
		adminFlags.Usage()
		return
	}

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
	systemUserid := db.GetSystemUserID()
	err = db.AddModerationLog(systemUserid, userid, constants.MODLOG_ADMIN_MAKE)
	if err != nil {
		complain("adding mod log for adding new admin failed (%w)", err)
	}

	inform("Successfully added %s (id %d) as an admin", username, userid)
	inform("Please visit %s for all your administration needs (changing usernames, resetting passwords, deleting user accounts)", adminRoute)
	inform("Admin action has been logged to /moderations")
}
