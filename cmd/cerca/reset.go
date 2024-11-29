package main

import (
	"cerca/constants"
	"cerca/database"
	"flag"
	"os"
)

func reset() {
	var username string
	var dbPath string

	resetCmd := flag.NewFlagSet("reset", flag.ExitOnError)
	resetCmd.StringVar(&username, "username", "", "username whose credentials should be reset")
	resetCmd.StringVar(&dbPath, "database", "./data/forum.db", "full path to the forum database; e.g. ./data/forum.db")

	help := createHelpString([]string{
		"cerca reset -username myCoolUsername",
	}, false)
	resetCmd.Usage = func() { usage(help, resetCmd) }
	resetCmd.Parse(os.Args[2:])

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
		complain("reset password failed (%w)", err)
	}
	newPassword, err := db.ResetPassword(userid)

	if err != nil {
		complain("reset password failed (%w)", err)
	}

	// log cmd actions just as admin web-actions are logged
	systemUserid := db.GetSystemUserid()
	err = db.AddModerationLog(systemUserid, userid, constants.MODLOG_RESETPW)
	if err != nil {
		complain("adding mod log for password reset failed (%w)", err)
	}

	inform("Successfully updated %s's password hash", username)
	inform("New temporary password: %s", newPassword)
	inform("Admin action has been logged to /moderations")
}
