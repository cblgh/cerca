package main

import (
	"flag"
	"gomod.cblgh.org/cerca/constants"
	"gomod.cblgh.org/cerca/database"
	"os"
)

func reset() {
	var username string
	var dbPath string

	resetFlags := flag.NewFlagSet("resetpw", flag.ExitOnError)
	resetFlags.StringVar(&username, "username", "", "username whose credentials should be reset")
	resetFlags.StringVar(&dbPath, "database", "", "full path to the forum database; e.g. ./data/forum.db")

	help := createHelpString("resetpw", []string{
		`cerca resetpw -username "<existing username>" -database "<path/to/forum.db>"`,
	})
	resetFlags.Usage = func() { usage(help, resetFlags) }
	resetFlags.Parse(os.Args[2:])

	// if run without flags, print the help info
	if resetFlags.NFlag() == 0 {
		resetFlags.Usage()
		return
	}

	if username == "" || dbPath == "" {
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
	systemUserid := db.GetSystemUserID()
	err = db.AddModerationLog(systemUserid, userid, constants.MODLOG_RESETPW)
	if err != nil {
		complain("adding mod log for password reset failed (%w)", err)
	}

	inform("Successfully updated %s's password hash", username)
	inform("New temporary password: %s", newPassword)
	inform("Admin action has been logged to /moderations")
}
